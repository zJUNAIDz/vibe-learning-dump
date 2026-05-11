
# 02-interview-questions-advanced.md

- **Purpose**: A collection of common advanced-level Git interview questions and ideal answers, focusing on internals, performance, and large-scale systems.

---

### 1. What are the four main types of objects in Git's object database?

- **Answer**: Git's object database primarily consists of four object types:
    1.  **Blob (Binary Large Object)**: This object stores the raw content of a file. It has no metadata, not even a filename.
    2.  **Tree**: This object represents a directory. It contains a list of pointers, where each pointer includes a mode, an object type (blob or tree), a SHA-1 hash, and a filename. It maps filenames to blobs and other trees (subdirectories).
    3.  **Commit**: This object points to a single tree object, representing the state of the repository at that point in time. It also contains pointers to one or more parent commits, creating the commit history graph. Additionally, it holds metadata like the author, committer, timestamp, and commit message.
    4.  **Tag**: This is a named pointer to a commit. An "annotated" tag is its own object type, containing a pointer to a commit, a tagger, a date, a message, and an optional GPG signature. A "lightweight" tag is just a ref (a file in `.git/refs/tags`) that points directly to a commit.

### 2. Explain the difference between a "fast-forward" merge and a "three-way" merge.

- **Answer**: The type of merge Git performs depends on the relationship between the two branches being merged.
    - A **fast-forward merge** can occur when the target branch has not diverged from the source branch. In other words, the tip of the target branch is a direct ancestor of the tip of the source branch. In this case, no new "merge commit" is needed. Git simply "fast-forwards" the target branch's pointer to point to the same commit as the source branch.
    - A **three-way merge** is necessary when the branches have diverged. Git uses three commits to resolve the merge: the two branch tips and their most recent common ancestor (the "merge base"). Git compares the changes between the common ancestor and each branch tip. It then combines these changes and creates a new "merge commit" that has both branch tips as its parents. This new commit represents the merged state.

### 3. What problem does `git sparse-checkout` solve? How does it work?

- **Answer**: `git sparse-checkout` solves the problem of working with massive monorepos where checking out the entire working directory would be incredibly slow and consume huge amounts of disk space. It allows a developer to specify a subset of the repository's directories that they actually need. When enabled, commands like `checkout` and `status` will only operate on the specified directories, ignoring the rest. It works by modifying the `read-tree` step. Instead of writing every file from the commit's tree object to the working directory, it filters the tree based on patterns defined in the `.git/info/sparse-checkout` file, only populating the working directory with the files and folders that match.

### 4. What is `git bisect` and how does it work?

- **Answer**: `git bisect` is a powerful debugging tool used to find the specific commit that introduced a bug. It works by performing a binary search on the commit history. The developer starts the process by providing a known "good" commit (where the bug is not present) and a known "bad" commit (where the bug is present). `git bisect` then automatically checks out a commit at the midpoint of this range and asks the developer to test it and label it as "good" or "bad". Based on the answer, it halves the search space and repeats the process until it has isolated the single commit that was the first to introduce the bug. The entire process can be automated if you can provide a script that can programmatically detect the bug.

### 5. What is the difference between `git push --force` and `git push --force-with-lease`?

- **Answer**: Both commands are used to forcibly overwrite the history of a remote branch, but they have a critical difference in safety.
    - `git push --force` is a blunt instrument. It tells the remote to unconditionally accept your local branch's history, overwriting whatever is on the remote. If a teammate has pushed new commits to the remote branch since you last fetched, `--force` will delete their work without warning.
    - `git push --force-with-lease` is a safer alternative. It's a conditional force push. It tells the remote, "I'm going to force push, but only if the remote branch still points to the commit I think it does." It checks that no one else has updated the remote branch in the meantime. If someone has, the push is rejected, preventing you from accidentally overwriting their work. It's the recommended way to perform any force push.

### 6. What is a "packfile"? What is the purpose of `git gc`?

- **Answer**: A packfile is a highly compressed archive file that contains multiple Git objects. When you first create objects (e.g., via a commit), Git often stores them as individual "loose" objects. This is inefficient for storage and network transfer. The `git gc` (garbage collection) command is a housekeeping tool that finds all these loose objects, delta-compresses them against each other to save space, and bundles them into a single packfile. It also cleans up and removes any "dangling" objects that are no longer reachable from any branch or tag, freeing up disk space.

### 7. What is Git LFS and what problem does it solve?

- **Answer**: Git LFS (Large File Storage) is a Git extension for versioning large binary files. Git itself is inefficient at handling large files because it stores the entire new version of a file for every change and its delta compression is not effective on binaries. This causes the repository to become bloated and slow. LFS solves this by storing the large files on a separate LFS server. Inside the Git repository, it replaces the large file with a small text "pointer file" that contains a hash and a URL for the actual file. When you `push`, the LFS client uploads the large file to the LFS server. When you `clone` or `pull`, the client reads the pointer file and downloads the large file from the LFS server. This keeps the core Git repository small and fast.

### 8. Explain the GitFlow branching model and discuss a scenario where it is a good choice, and a scenario where it is a poor choice.

- **Answer**: GitFlow is a structured branching model with two long-lived branches, `main` (for production releases) and `develop` (for integration). Features are developed on `feature` branches off of `develop`. When a release is planned, a `release` branch is created from `develop` for stabilization. Once ready, it's merged into `main` and `develop`. Emergency production fixes are done on `hotfix` branches off of `main` and also merged back to `develop`.
    - **Good Choice**: GitFlow is an excellent choice for software that has scheduled, versioned releases, like a desktop application, a mobile app, or an API with a versioned contract. The structure provides a high degree of stability and a clear process for managing releases.
    - **Poor Choice**: It's a poor choice for a web application that practices continuous deployment. The overhead of the multiple branches and merges slows down the delivery process significantly. For a team that wants to ship multiple times a day, a simpler model like GitHub Flow or Trunk-Based Development is far more suitable.

### 9. What are Git "plumbing" vs. "porcelain" commands? Give an example of each.

- **Answer**: This is a metaphor used to describe the two layers of Git commands.
    - **Porcelain** commands are the high-level, user-facing commands that are easy to use and understand. They are the "shiny interface." Examples include `git log`, `git status`, and `git commit`.
    - **Plumbing** commands are the low-level commands that do the real work behind the scenes. They are designed to be chained together in scripts and are not typically used directly by humans. They are the "pipes" that make the porcelain work. Examples include `git hash-object` (creates a blob), `git cat-file` (inspects an object), `git write-tree` (creates a tree), and `git commit-tree` (creates a commit). Understanding plumbing is key to understanding Git's internals.

### 10. How would you go about removing a large, sensitive file that was accidentally committed and now exists deep in the repository's history? What are the consequences of this action?

- **Answer**: The correct tool for this is `git filter-repo`, which is the modern replacement for the older, more dangerous `git filter-branch`. The process is:
    1.  First, and most importantly, treat the sensitive data as compromised and rotate any credentials.
    2.  Communicate to the entire team that a history rewrite is about to happen. This is a disruptive, "break-glass" procedure.
    3.  Create a fresh, mirror clone of the repository as a backup.
    4.  Run `git filter-repo --path path/to/the/file --invert-paths` to remove the file from every commit in the repository's history.
    5.  Force-push the new, rewritten history to the remote repository, overwriting the old history: `git push origin --force --all`.
    - **Consequences**: This action rewrites the entire commit history, changing the SHA-1 of every single commit. This means that all existing forks, clones, and pull requests will be invalidated. Every developer on the team must delete their old local copy and re-clone the repository from the remote to continue working. It's a highly disruptive but necessary action for security incidents.
