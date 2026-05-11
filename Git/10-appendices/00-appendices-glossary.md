
# 00-appendices-glossary.md

- **Purpose**: A glossary of common and advanced Git terms used throughout this course.

---

### A

- **Alias**: A custom shortcut for a longer Git command, configured in `.gitconfig`.
- **Ancestor**: A commit that comes before another commit in its history.
- **Amend**: To edit the most recent commit (`git commit --amend`). This doesn't actually "edit" the commit; it creates a new commit that replaces the old one.

### B

- **Bare Repository**: A Git repository that does not have a working directory. It only contains the `.git` data. These are typically used as central remote repositories.
- **Bisect**: A command (`git bisect`) that performs a binary search on the commit history to find the specific commit that introduced a bug.
- **Blame**: A command (`git blame`) that annotates each line of a file to show who last modified it and in which commit.
- **Blob**: A "binary large object." In Git's object database, a blob represents the content of a file.
- **Branch**: A lightweight, movable pointer to a commit.

### C

- **Cherry-pick**: The action of taking a single commit from one branch and applying it as a new commit on another branch (`git cherry-pick`).
- **Checkout**: The action of switching `HEAD` to point to a different branch or commit. Also, historically, the command used to restore files. See `switch` and `restore`.
- **Clone**: To create a local copy of a remote repository.
- **Commit (object)**: A snapshot of the repository at a specific point in time. A commit object contains a pointer to a tree object, parent commit(s), and metadata like the author, committer, date, and message.
- **Commit-graph**: A supplemental file that acts as an index of the commit history, allowing Git to traverse the graph much more quickly.
- **Conventional Commits**: A specification for formatting commit messages that makes them machine-readable and helps automate changelog generation.
- **CI/CD**: Continuous Integration and Continuous Deployment. A set of practices for automating the building, testing, and deployment of software.

### D

- **Dangling Object**: A commit or other object that is not reachable from any branch or tag. These are eventually removed by garbage collection.
- **Detached HEAD**: A state where `HEAD` points directly to a commit SHA instead of to a branch name. This happens when you `checkout` a commit or a tag.
- **Diff**: The difference in content between two files or two commits.
- **`develop` branch**: The main integration branch in the GitFlow branching model.

### F

- **Fast-forward**: A type of merge where the target branch's tip is a direct ancestor of the source branch's tip. No merge commit is needed; Git just moves the branch pointer forward.
- **Fetch**: To download objects and refs from a remote repository without merging them into your local branches (`git fetch`).
- **Filter-repo**: The modern, recommended tool for rewriting repository history (e.g., to remove a file from all commits).
- **Force Push**: A push that overwrites the history on the remote branch (`git push --force`). It's dangerous and should be used with caution, preferably with `--force-with-lease`.

### G

- **Garbage Collection (`gc`)**: A Git process that cleans up unreachable objects and packs loose objects into more efficient packfiles.
- **GitFlow**: A structured branching model with long-lived `main` and `develop` branches, suitable for projects with scheduled releases.
- **GitHub Flow**: A lightweight branching model where `main` is always deployable and features are developed on short-lived branches and merged via pull requests.

### H

- **HEAD**: A special pointer that indicates your current location in the repository. It usually points to a local branch name (e.g., `main`), but can also point directly to a commit SHA (in a "detached HEAD" state).
- **Hook**: A script that Git executes automatically before or after a specific event (e.g., `pre-commit`, `pre-push`).

### I

- **Index**: Also known as the "staging area." A file in the `.git` directory that stores information about what will go into your next commit. It sits between your working directory and your commit history.

### L

- **LFS (Large File Storage)**: A Git extension for versioning large binary files. It stores the large files on a separate server and keeps lightweight pointers in the Git repository.

### M

- **`main` branch**: The default branch name in modern Git. It typically represents the primary line of development or the stable, production-ready history.
- **Merge**: To combine the history of two or more branches.
- **Merge Commit**: A commit that has two or more parents, created by a `git merge` operation.
- **Monorepo**: A repository strategy where all of an organization's code lives in a single Git repository.

### O

- **Object Database**: The core of Git, located in the `.git/objects` directory. It stores all commit, tree, blob, and tag objects.
- **`origin`**: The default name given to the remote repository you cloned from.
- **`ORIG_HEAD`**: A special pointer that Git creates during dangerous operations like `merge` or `rebase` to store the location of `HEAD` before the operation started. It's a safety net.

### P

- **Packfile**: A highly compressed archive file containing multiple Git objects. `git gc` is the command that creates packfiles.
- **Plumbing**: Low-level Git commands that are designed to be used in scripts (e.g., `hash-object`, `cat-file`).
- **Polyrepo**: A repository strategy where each project or service has its own separate repository.
- **Porcelain**: High-level Git commands that are designed for human use (e.g., `log`, `status`, `commit`).
- **Pull**: The action of fetching from a remote and immediately merging the changes (`git pull`). It's equivalent to `git fetch` + `git merge`.
- **Pull Request (PR)**: A feature of Git hosting platforms (like GitHub) that provides a forum for reviewing and discussing a proposed change before it is merged.
- **Push**: To upload your local commits to a remote repository (`git push`).

### R

- **Rebase**: The action of re-applying a series of commits on top of a different base commit (`git rebase`). It rewrites history.
- **Ref (Reference)**: A pointer to a commit. Branches and tags are refs. They are stored in the `.git/refs` directory.
- **Reflog**: A local-only log of all the movements of `HEAD` and branch tips. It's the ultimate safety net for recovering "lost" work.
- **Remote**: A named pointer to another Git repository.
- **Remote-tracking branch**: A local, read-only pointer to the state of a branch on a remote repository (e.g., `origin/main`).
- **Reset**: A command (`git reset`) to move the current branch's tip to a different commit, optionally modifying the index and working directory.
- **Restore**: A modern command (`git restore`) used to discard changes in the working directory or to unstage files.

### S

- **SHA (Secure Hash Algorithm)**: The algorithm used to generate the unique 40-character ID for every object in Git.
- **Shallow Clone**: A clone of a repository that only includes a limited number of recent commits, not the entire history.
- **Sparse-checkout**: A Git feature that allows you to check out only a subset of the files in a repository, which is useful for very large monorepos.
- **Squash**: To combine multiple commits into a single commit, typically done during an interactive rebase.
- **Staging Area**: See `Index`.
- **Stash**: A temporary storage area for uncommitted changes (`git stash`).
- **Switch**: A modern command (`git switch`) used to change branches.

### T

- **Tag**: A pointer to a specific commit, typically used to mark a release version (e.g., `v1.0`).
- **Tree (object)**: A Git object that represents a directory. It contains pointers to blobs (files) and other trees (subdirectories).
- **Trunk-Based Development (TBD)**: A branching model where all developers commit to a single `main` branch (the "trunk").

### W

- **Working Directory**: The directory of files on your local machine that you are actively editing.
- **Worktree**: A feature (`git worktree`) that allows you to have multiple working directories, each on a different branch, attached to a single Git repository.
