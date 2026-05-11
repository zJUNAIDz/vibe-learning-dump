
# 02-optimizing-git-performance.md

- **Purpose**: To provide a collection of practical tips and configuration settings for improving Git's performance on large repositories.
- **Estimated Difficulty**: 4/5
- **Estimated Reading Time**: 40 minutes
- **Prerequisites**: `01-git-at-scale-vfs-and-sparse-checkout.md`

---

### Why Does Git Get Slow?

Git is incredibly fast on small to medium-sized repositories. However, performance can degrade as a repository grows in size (GBs of data) and history (millions of commits or objects).

Common performance bottlenecks include:
- **Large number of objects**: `git gc` (garbage collection) can take a long time.
- **Large number of refs**: Having tens of thousands of branches or tags can slow down operations that need to scan them.
- **Large files**: Storing large binary assets in Git can bloat the repository size.
- **Working directory size**: Commands like `git status` need to scan every file in the working directory, which is slow if there are millions of files.

This lesson focuses on configuration and commands to mitigate these issues.

### 1. Garbage Collection (`git gc`)

Git stores objects in "loose" files. Over time, this is inefficient. `git gc` is a housekeeping command that packs these loose objects into highly compressed "packfiles" and removes unreachable objects.

Git runs `gc` automatically, but sometimes you need to tune it.

- **Aggressive GC**: `git gc --aggressive` spends more time finding better compression deltas. It's slower but can result in a smaller repository. Use this sparingly.
- **Pruning**: `git gc --prune=now` will immediately expire and remove any unreachable objects (e.g., from old rebases). This can recover disk space but also removes your reflog-based safety net for those objects.

**The Commit-Graph**:
A major performance boost for history-traversal operations (`log`, `blame`, `bisect`) is the commit-graph. This is a pre-computed index of the commit history.

You can enable it and update it with:
```bash
# Enable it for the current repo
$ git config core.commitGraph true

# Generate or update the commit-graph file
$ git commit-graph write --reachable
```
For very large repos, this can make `git log` orders of magnitude faster.

### 2. The Filesystem Monitor (FSMON)

`git status` can be slow because it has to `lstat()` every file in your working directory to see if it has changed. The Filesystem Monitor (or "watchman") integrates Git with your operating system's file-watching capabilities.

When enabled, Git can ask the OS "what files have changed since the last time I asked?" instead of scanning everything itself. This can make `git status` nearly instantaneous, even on huge working directories.

**How to enable it:**
```bash
# For the current repo
$ git config core.fsmonitor true
```
This requires a recent version of Git and may need some OS-level setup. It's a massive performance win for large projects.

### 3. Partial Clones (`--filter`)

When you clone a repository, you don't always need the full history of every large file. A partial clone lets you clone a repository without downloading blob (file content) objects.

```bash
$ git clone --filter=blob:none <repo_url>
```
This will download all the commit and tree objects, so you have the full history, but it will not download any file contents. The file blobs will be downloaded from the remote on-demand when you check out a branch and Git needs them.

This is a great way to quickly clone a large repository without downloading gigabytes of binary assets you may not need.

### 4. Shallow Clones (`--depth`)

If you don't need the full history of a project, you can perform a shallow clone. This is common in CI/CD environments where you only need the latest version of the code to run a build.

```bash
# Clone only the most recent commit
$ git clone --depth=1 <repo_url>
```
This creates a clone with a truncated history. This is very fast but has limitations. You cannot, for example, push from a shallow clone back to the original repository without doing extra work, because the histories don't match.

### 5. Git LFS (Large File Storage)

This is the standard solution for handling large binary files (e.g., images, videos, compiled libraries) in Git.

**Problem**: Git is designed for text. When you change a large binary file, Git stores the *entire new file*, even if only one byte changed. This causes the repository to bloat very quickly.

**Solution**: `git lfs` works by storing the large files on a separate LFS server. In your Git repository, instead of the large file itself, it stores a tiny "pointer file" that contains the SHA-256 hash and size of the large file.

**Workflow:**
1.  `git lfs install` (one-time setup).
2.  `git lfs track "*.psd"`: Tell LFS to handle all Photoshop files. This creates/updates the `.gitattributes` file.
3.  You `git add`, `commit`, and `push` as normal.
4.  Behind the scenes, `git lfs` intercepts the large files, uploads them to the LFS server, and commits the small pointer files to your Git repository.

When someone else clones the repo and runs `git lfs pull`, LFS reads the pointer files and downloads the actual large files from the LFS server.

This keeps your Git repository small and fast, while still versioning your large assets.

### Summary of Performance Configurations

| Tactic                  | Command / Config                               | Problem It Solves                                     |
| ----------------------- | ---------------------------------------------- | ----------------------------------------------------- |
| **Commit Graph**        | `git config core.commitGraph true`             | Slow history traversal (`log`, `blame`).              |
| **Filesystem Monitor**  | `git config core.fsmonitor true`               | Slow `git status` in large working directories.       |
| **Partial Clone**       | `git clone --filter=blob:none`                 | Slow initial clone due to large file history.         |
| **Shallow Clone**       | `git clone --depth=1`                          | Slow initial clone due to long commit history.        |
| **Sparse Checkout**     | `git sparse-checkout set <dirs>`               | Large working directory with many unnecessary files.  |
| **Git LFS**             | `git lfs track "*.bin"`                        | Repository bloat from large binary files.             |

By combining these techniques, you can make working with even the largest repositories a fast and pleasant experience.
