
# 00-interview-questions-beginner.md

- **Purpose**: A collection of common beginner-level Git interview questions and ideal answers.

---

### 1. What is Git? What is it used for?

- **Answer**: Git is a distributed version control system (DVCS). It's used to track changes in source code during software development. It allows multiple developers to collaborate on a project by keeping a full history of all changes, managing different versions of the code, and helping to merge work from different people. Its key features are that it's distributed (everyone has a full copy of the repository's history), it's fast, and it's excellent at branching and merging.

### 2. What is the difference between Git and GitHub?

- **Answer**: Git is the tool, the version control system itself. It's a command-line program that runs on your local machine. GitHub is a web-based hosting service for Git repositories. It provides a central place to store your repositories remotely, and it adds a layer of collaboration tools on top of Git, such as pull requests, issue tracking, and project management boards. So, Git is the tool, and GitHub is a service that hosts projects that use Git.

### 3. What is a "commit"?

- **Answer**: A commit is a snapshot of your repository at a specific point in time. When you run `git commit`, you are saving the current state of your staged files. Each commit has a unique SHA-1 ID, a commit message, and pointers to its parent commits, which creates the project history.

### 4. What is the "staging area" or "index"?

- **Answer**: The staging area is an intermediate step between your working directory and your commit history. It's where you prepare, or "stage," the changes that you want to include in your next commit. This allows you to be selective and craft a clean, atomic commit that only contains related changes, even if you have many other unrelated modifications in your working directory. You use `git add` to move changes to the staging area.

### 5. What is the difference between `git fetch` and `git pull`?

- **Answer**: `git fetch` downloads the latest changes, commits, and branches from a remote repository but does *not* merge them into your local working branch. It just updates your remote-tracking branches (like `origin/main`). `git pull` does a `git fetch` and then immediately tries to merge the downloaded changes into your current local branch. `fetch` is a safe, read-only operation, while `pull` is a fetch followed by a potentially state-changing merge.

### 6. What is a "merge conflict"? How do you resolve it?

- **Answer**: A merge conflict occurs when Git is unable to automatically merge changes from two different branches because they have made competing changes to the same lines in the same file. Git will stop the merge and insert conflict markers (`<<<<<<<`, `=======`, `>>>>>>>`) into the problematic file. To resolve it, you must:
    1.  Open the conflicted file and manually edit it to the final, correct state, removing the conflict markers.
    2.  Use `git add <file>` to tell Git that you have resolved the conflict in that file.
    3.  Once all conflicts are resolved and staged, run `git commit` to finalize the merge.

### 7. What is a "branch"? Why is branching useful?

- **Answer**: A branch is a lightweight, movable pointer to a commit. It represents an independent line of development. Branching is useful because it allows you to work on a new feature or a bug fix in an isolated environment without affecting the main, stable codebase (the `main` branch). Once your work is complete and tested on the branch, you can merge it back into the main branch. This allows for parallel development and keeps the main line of code clean.

### 8. How would you revert a commit that has already been pushed and is public?

- **Answer**: You should not use `git reset` to remove a public commit, as that rewrites history. The safe way to "undo" a public commit is to use `git revert <commit-sha>`. This command creates a *new* commit that applies the inverse of the changes from the specified commit. This doesn't rewrite history; it moves the history forward by explicitly undoing the previous work, which is a safe operation to share with others.

### 9. What is the difference between `HEAD`, the working directory, and the index?

- **Answer**: These are the "three trees" of Git.
    - The **working directory** is the set of files on your local filesystem that you are currently editing.
    - The **index** (or staging area) is where you place files you want to include in your next commit. It's a snapshot you are preparing.
    - **`HEAD`** is a pointer to the last commit on your current branch. It represents the last known-good state of your repository.
    The basic workflow is to make changes in your working directory, stage them to the index with `git add`, and then save the index as a permanent snapshot with `git commit`, which `HEAD` then points to.

### 10. What is a `.gitignore` file?

- **Answer**: The `.gitignore` file is a text file where you can list files or directories that you want Git to ignore. Git will not track changes to these files, and they won't show up in `git status` or be included when you run `git add .`. This is used for files that are specific to a local environment, like build artifacts, log files, dependency directories (`node_modules`), or files containing secrets.
