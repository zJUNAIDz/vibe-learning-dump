
# 07-lab-build-a-commit-from-scratch.md

- **Purpose**: A hands-on lab to solidify understanding of the object model by manually creating a commit using only plumbing commands.
- **Estimated Difficulty**: 5/5
- **Estimated Time**: 60 minutes
- **Prerequisites**: All previous lessons in Module 01.

---

### Goal

In this lab, we will perform the actions of `git add` and `git commit` without using those commands. We will use only the low-level plumbing commands we've learned about to construct a commit from scratch. This will prove our understanding of blobs, trees, and commits.

### Setup

Start with a fresh, empty repository.

```bash
$ mkdir manual-commit-lab && cd manual-commit-lab
$ git init
```

### Step 1: Create Content and Blob Objects

First, let's create the file structure for our project.

```bash
$ echo 'Hello, Git!' > README.md
$ mkdir src
$ echo 'const main = () => "Hello, World!";' > src/index.js
```

Now, we need to create blob objects for each of these files and record their SHAs.

```bash
$ readme_blob_sha=$(git hash-object -w README.md)
$ index_blob_sha=$(git hash-object -w src/index.js)

$ echo $readme_blob_sha
# e.g., 089e5b69e15920d16c343a5197f7499823e50111

$ echo $index_blob_sha
# e.g., 1853b7a95f545183d37883764343edc81c708da0
```
*(Your SHAs will be different if your content or Git version differs slightly)*

### Step 2: Create the Tree Objects

We have two directories: the root (`/`) and `src/`. This means we need two tree objects. We'll start from the bottom up.

First, create the tree for the `src` directory. It contains one file, `index.js`. The `git mktree` command can read tree information from standard input. The format is `<mode> <type> <sha>\t<name>`.

```bash
$ src_tree_sha=$(printf "100644 blob %s\tindex.js" "$index_blob_sha" | git mktree)

$ echo $src_tree_sha
# e.g., 8b726a13b84e7036c359f09394f4a56238699b34
```

Now, let's inspect it to be sure:
```bash
$ git cat-file -p $src_tree_sha
100644 blob 1853b7a95f545183d37883764343edc81c708da0    index.js
```
Perfect. Now we create the root tree. It contains `README.md` and the `src` directory.

```bash
$ root_tree_sha=$(printf "100644 blob %s\tREADME.md\n040000 tree %s\tsrc" "$readme_blob_sha" "$src_tree_sha" | git mktree)

$ echo $root_tree_sha
# e.g., 2b421835a052a0517387343812e313abff5f225c
```

Let's inspect the root tree:
```bash
$ git cat-file -p $root_tree_sha
100644 blob 089e5b69e15920d16c343a5197f7499823e50111    README.md
040000 tree 8b726a13b84e7036c359f09394f4a56238699b34    src
```
Our directory structure is now correctly represented in the object database.

### Step 3: Create the Commit Object

We have our top-level tree. Now we can create the commit. Since this is the first commit, it has no parent.

```bash
$ commit_message="Feat: Initial commit, created manually"
$ commit_sha=$(echo "$commit_message" | git commit-tree $root_tree_sha)

$ echo $commit_sha
# e.g., 4a2b8b74b78d6c3e421a3e5d6f7c8d9a0b1c2d3e
```

Let's inspect our final commit object:
```bash
$ git cat-file -p $commit_sha
tree 2b421835a052a0517387343812e313abff5f225c
author Your Name <you@example.com> ...
committer Your Name <you@example.com> ...

Feat: Initial commit, created manually
```

### Step 4: Point a Branch at the New Commit

The commit exists, but it's unreachable. `git log` will show nothing. We need to update a branch ref to point to our new commit. Let's update `master`.

```bash
$ git update-ref refs/heads/master $commit_sha
```

### Verification

Now, let's use the normal "porcelain" commands to see if our work was successful.

```bash
$ git log
commit 4a2b8b74b78d6c3e421a3e5d6f7c8d9a0b1c2d3e (HEAD -> master)
Author: Your Name <you@example.com>
Date:   ...

    Feat: Initial commit, created manually

$ git ls-files
README.md
src/index.js

$ cat README.md
Hello, Git!
```

It worked! We have successfully created a commit from scratch, demonstrating a complete understanding of the blob -> tree -> commit relationship.

### Challenge

Create a *second* commit manually. This commit should:
1.  Modify the content of `README.md`.
2.  Use the first commit we created as its parent.
3.  Update the `master` branch to point to this new commit.

This will require you to:
- Create a new blob for the modified `README.md`.
- Create a new root tree that points to the new blob but re-uses the existing `src` tree (since it didn't change).
- Use `git commit-tree` with the `-p` flag to specify the parent commit.
- Update the `master` ref again.
