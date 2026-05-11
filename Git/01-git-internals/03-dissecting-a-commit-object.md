
# 03-dissecting-a-commit-object.md

- **Purpose**: To use plumbing commands to create and inspect a raw commit object.
- **Estimated Difficulty**: 4/5
- **Estimated Reading Time**: 35 minutes
- **Prerequisites**: `02-dissecting-a-tree-object.md`

---

### What is a Commit?

A commit is the object that ties everything together into a historical snapshot. It's a simple text file that contains:

- The SHA-1 of the **top-level tree** for this snapshot.
- The SHA-1 of the **parent commit(s)**.
- Author and Committer information.
- A commit message.

### Lab: Creating Our First Commit Manually

We have a tree object from the last lesson. Now let's wrap it in a commit.

**1. Get the Tree SHA**
First, let's get our tree SHA again. If you've followed along, your working directory and index should be in the same state.

```bash
# In git-internals-lab
$ git write-tree
f6621c2ab78345553439973338c68334d2598913
```
Let's store this in a variable for convenience.
```bash
$ tree_sha=$(git write-tree)
$ echo $tree_sha
f6621c2ab78345553439973338c68334d2598913
```

**2. Create the Commit Object**
The `git commit-tree` command creates a commit object. It takes a tree SHA and, optionally, one or more parent commit SHAs. Since this is our first commit, it has no parent.

We'll pipe the commit message to the command via standard input.

```bash
$ echo "Initial commit" | git commit-tree $tree_sha
c1f86f1554c52b55a67f42e3a32589510a185997
```
This returns the SHA-1 of our newly created commit object.

**3. Inspect the Commit Object**
Let's use `cat-file` to see what we've made.

```bash
$ git cat-file -p c1f86f1554c52b55a67f42e3a32589510a185997
tree f6621c2ab78345553439973338c68334d2598913
author Your Name <you@example.com> 1678886400 -0700
committer Your Name <you@example.com> 1678886400 -0700

Initial commit
```
It's all there!
- The `tree` pointer matches the one we provided.
- There is no `parent` line, because this is the root commit.
- The author/committer info is pulled from your Git config.
- The commit message is what we piped in.

**4. Where is this commit?**
If you run `git log`, you won't see this commit. Why? Because no **ref** (like a branch) is pointing to it. It's an "unreachable" object. It exists in the object database, but nothing is referencing it.

```bash
$ git log
fatal: your current branch 'master' does not have any commits yet
```

To make it visible, we need to attach a branch to it.

```bash
$ git update-ref refs/heads/master c1f86f1554c52b55a67f42e3a32589510a185997
```
This command directly manipulates the `master` branch pointer to point to our new commit SHA. Now let's check the log.

```bash
$ git log
commit c1f86f1554c52b55a67f42e3a32589510a185997 (HEAD -> master)
Author: Your Name <you@example.com>
Date:   ...

    Initial commit
```
Success! We have manually created a commit and made it the tip of our `master` branch. This is exactly what `git commit` does under the hood: it runs `write-tree` and then `commit-tree`.

### Lab: Creating a Second Commit

Let's make a second commit that points to our first one as its parent.

```bash
# 1. Make a change and update the index
$ echo "version 2" > file1.txt
$ git add file1.txt

# 2. Create a new tree for this state
$ new_tree_sha=$(git write-tree)

# 3. Get the parent commit's SHA
$ parent_sha=$(git rev-parse HEAD)

# 4. Create the new commit, specifying the parent with -p
$ echo "Second commit" | git commit-tree $new_tree_sha -p $parent_sha
a8a5e4a9c29e136edc5f8789f6d3f80833c45334

# 5. Update the master branch to point to the new commit
$ git update-ref refs/heads/master a8a5e4a9c29e136edc5f8789f6d3f80833c45334
```

Now, let's inspect the new commit and the log:
```bash
$ git cat-file -p a8a5e4a9c29e136edc5f8789f6d3f80833c45334
tree ...
parent c1f86f1554c52b55a67f42e3a32589510a185997
author ...

Second commit

$ git log --oneline --graph
* a8a5e4a (HEAD -> master) Second commit
* c1f86f1 SInitial commit
```
We have successfully built a chain of history by manually creating commit objects and linking them via the `parent` pointer.

### Key Takeaways

- A commit is a snapshot pointing to a tree and its parent(s).
- `git commit-tree` is the low-level command for creating a commit object.
- A commit is unreachable until a ref (like a branch) points to it.
- The chain of parent pointers forms the directed acyclic graph (DAG) of your project's history.

### Interview Notes

- **Question**: "Walk me through what happens when you run `git commit`."
- **Answer**: "First, `git commit` creates a tree object that represents the current state of the index. This is like running `git write-tree`. Then, it creates a new commit object using `git commit-tree`. This new object points to the tree that was just created and also points to the current HEAD commit as its parent. Finally, it updates the ref for the current branch (e.g., `refs/heads/master`) to point to this new commit object, making it the new HEAD."
