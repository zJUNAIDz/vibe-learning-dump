
# 01-dissecting-a-blob-object.md

- **Purpose**: To use plumbing commands to create and inspect a raw blob object.
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 30 minutes
- **Prerequisites**: `01-git-internals/00-the-object-database.md`

---

### What is a Blob?

A blob is the simplest object in Git. It's just the raw content of a file. It has no name, no permissions, nothing but the data itself. Let's prove this by creating one from scratch.

### Lab: Creating and Inspecting a Blob

**1. Setup**
First, let's create a temporary, empty Git repository to play in. This is a safe space where we can't break anything important.

```bash
$ mkdir git-internals-lab
$ cd git-internals-lab
$ git init
Initialized empty Git repository in /path/to/git-internals-lab/.git/
```

**2. Create Content**
Let's create a simple file.

```bash
$ echo 'hello world' > hello.txt
```

**3. Create a Blob Object**
The `git hash-object` command takes a file, computes its hash, and, with the `-w` flag, writes it to the object database.

```bash
$ git hash-object -w hello.txt
d9014c4f2f5b1a24350b5939f37c3a564288a47f
```

This 40-character string is the SHA-1 hash of our blob object. Git has just created a new object in `.git/objects`. Let's look.

```bash
$ find .git/objects -type f
.git/objects/d9/014c4f2f5b1a24350b5939f37c3a564288a47f
```
Git stores objects in a directory named with the first two characters of the hash, and the file is named with the remaining 38 characters. This is a simple optimization to prevent having too many files in one directory.

**4. Inspect the Blob Object**
Now, let's use another plumbing command, `git cat-file`, to inspect this object.

- `cat-file -t` shows the object's type.
- `cat-file -p` "pretty-prints" the object's content.

```bash
$ git cat-file -t d9014c4f2f5b1a24350b5939f37c3a564288a47f
blob

$ git cat-file -p d9014c4f2f5b1a24350b5939f37c3a564288a47f
hello world
```
This confirms it. The object with that SHA-1 is a `blob`, and its content is exactly `hello world`. Notice that the filename `hello.txt` is nowhere to be found. The blob only knows about the content.

### How the Hash is Calculated

The SHA-1 hash isn't just of the raw file data. It's calculated on the object's header plus its content. For a blob, the header is:

`"blob" + " " + <content_length_in_bytes> + "\0" + <content>`

Let's verify this manually.

```bash
# The content is 'hello world' followed by a newline, which is 12 bytes.
$ printf "blob 12\0hello world\n"
blob 12hello world

# Pipe that through the sha1sum utility
$ printf "blob 12\0hello world\n" | sha1sum
d9014c4f2f5b1a24350b5939f37c3a564288a47f  -
```
It matches! This demonstrates the deterministic, content-addressed nature of Git.

### Key Takeaways

- A blob is pure content.
- The filename is not part of the blob object.
- The SHA-1 hash is derived from the object's header and its content.
- You can create and inspect objects manually using `hash-object` and `cat-file`.

### Exercises

1.  Create a new file with different content and find its hash using `git hash-object -w`.
2.  Verify that two identical files in different locations produce the exact same blob object hash.
3.  What happens if you create a file, hash it, then delete the file? Does the object remain in the `.git/objects` directory? (Hint: Yes. Why?)

### Interview Notes

- **Question**: "If I have two files with the same content but different names in my repository, how does Git store them?"
- **Answer**: "Git will only create a single blob object because the content is identical. The blob is content-addressed, not name-addressed. The different filenames will be handled by the tree object(s) that point to this single blob, effectively de-duplicating the content at the storage level."
