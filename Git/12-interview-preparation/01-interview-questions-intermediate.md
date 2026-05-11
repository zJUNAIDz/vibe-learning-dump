
# 01-interview-questions-intermediate.md

- **Purpose**: A collection of common intermediate-level Git interview questions and ideal answers, focusing on workflows and collaboration.

---

### 1. What is the difference between `git merge` and `git rebase`?

- **Answer**: Both `merge` and `rebase` are used to integrate changes from one branch into another, but they do so in different ways.
    - `git merge` creates a new "merge commit" that has two parents, tying the two histories together. It preserves the exact history of the branches. The resulting graph is a more accurate but often noisier representation of what happened.
    - `git rebase` takes the commits from one branch and "replays" them, one by one, on top of another branch. This rewrites history by creating new commits, resulting in a clean, linear history.
    - You would use `merge` when you want to preserve the historical context of a feature branch. You would use `rebase` to clean up your local history before sharing it, or to update your feature branch with the latest changes from `main`.

### 2. What is `git reset --hard`, `git reset --soft`, and `git reset --mixed`?

- **Answer**: `git reset` moves the `HEAD` pointer of the current branch to a specified commit. The flags determine how it affects the index and the working directory.
    - `--soft`: Moves `HEAD` only. The index and working directory are left untouched. The changes from the "undone" commits are left staged. Use case: To re-do a commit with a new message or combine it with other changes.
    - `--mixed` (the default): Moves `HEAD` and updates the index to match `HEAD`. The working directory is untouched. The changes are kept as unstaged modifications. Use case: To completely re-think your last commit and what you want to stage.
    - `--hard`: Moves `HEAD` and updates both the index and the working directory to match `HEAD`. This is destructive, as it discards all staged and unstaged changes. Use case: To completely throw away work and start fresh from a specific commit.

### 3. What is the "reflog"? How is it useful?

- **Answer**: The reflog is a local-only log that records almost every change made to the tips of branches and `HEAD`. It's Git's safety net. It's useful for recovering from mistakes. For example, if you accidentally `git reset --hard` and lose a commit, or delete a branch prematurely, the commit is not truly gone. You can use `git reflog` to find the SHA of the "lost" commit and then use `git reset` or `git checkout` to restore it.

### 4. What is `git cherry-pick` and when would you use it?

- **Answer**: `git cherry-pick` is a command that takes a single commit from one branch and applies it as a new commit on the current branch. You would use it when you need a specific commit from another branch, but you don't want to merge the entire branch. A classic use case is backporting a hotfix. If you fix a critical bug on `main`, but also need that same fix on an older, supported `release` branch, you can `cherry-pick` the fix commit onto the release branch without bringing in any other new features.

### 5. What is a "detached HEAD"? What does it mean?

- **Answer**: A detached HEAD is a state where `HEAD` is pointing directly to a commit SHA, rather than to a branch name. This happens if you `git checkout <commit-sha>` or `git checkout <tag-name>`. It means you are no longer on a branch. You can look around and make experimental commits, but these new commits don't belong to any branch. If you then switch back to a branch, those experimental commits will be "lost" (though they are still in the reflog for a while). If you want to keep them, you need to create a new branch from that point using `git switch -c <new-branch-name>`.

### 6. What is the difference between a remote-tracking branch (like `origin/main`) and a local branch (like `main`)?

- **Answer**: A local branch, `main`, is your own private line of development. You can commit to it, reset it, and rebase it. It's a pointer that you control. A remote-tracking branch, `origin/main`, is a local, read-only pointer that reflects the state of the `main` branch on the remote named `origin` the last time you ran `git fetch`. Its purpose is to be a local bookmark of the remote's state, so you can work offline and compare your work to the remote's work without constantly connecting to the network. You cannot commit directly to `origin/main`; it is only moved by `git fetch`.

### 7. You've just run `git push -f` on `main` by accident. How do you recover?

- **Answer**: This is an emergency that requires immediate, coordinated action.
    1.  First, lock the `main` branch on the remote (e.g., in GitHub's settings) to prevent anyone else from pushing and making the situation worse.
    2.  The goal is to find the SHA of the commit that `main` *should* be pointing to. The best place to find this is on a teammate's machine who has not yet pulled the bad changes. Their local `main` is the source of truth. Other places to look are CI server logs or the remote's activity feed.
    3.  Once you have the correct SHA, one person should fix it locally: `git switch main` followed by `git reset --hard <correct-sha>`.
    4.  That person then force-pushes the corrected branch to the remote: `git push --force-with-lease origin main`.
    5.  Finally, communicate to the entire team that they must update their local `main` branches by running `git reset --hard origin/main`, not by pulling or merging.

### 8. What is a "squash" commit? Why is it useful?

- **Answer**: Squashing is the act of combining multiple commits into a single one. This is typically done during an interactive rebase (`git rebase -i`). It's useful for cleaning up the history of a feature branch before it's merged. A feature branch might have many messy "work-in-progress" or "fix typo" commits. Before submitting a pull request, you can squash these down into a single, clean, well-documented commit that describes the entire feature. This keeps the `main` branch history clean and easy to read. Many teams use the "Squash and Merge" button on GitHub to automate this process.

### 9. What is the purpose of the `.gitattributes` file?

- **Answer**: `.gitattributes` is a file that allows you to specify attributes for certain paths or file patterns in your repository. Its most common uses are:
    - **Managing line endings**: You can specify whether text files should use `LF` or `CRLF` line endings, which prevents issues when developers are on different operating systems.
    - **Configuring Git LFS**: It's where you define which file patterns (e.g., `*.psd`) should be handled by Git LFS instead of being stored directly in Git.
    - **Defining diff strategies**: You can tell Git how to generate diffs for certain file types, for example, how to show a "diff" for an image file by showing its dimensions.

### 10. What is a "submodule"? What are its pros and cons?

- **Answer**: A Git submodule is a feature that allows you to embed one Git repository inside another. The outer repository stores a reference to a specific commit in the inner repository.
    - **Pros**: It allows you to include a dependency from another project (like a third-party library) while keeping its commit history separate from your own. You can lock the dependency to a specific version (commit).
    - **Cons**: They can be complex to work with. Developers on your team need to run extra commands (`git submodule init` and `git submodule update`) to get the submodule code. Updating a submodule to a new version is a manual process. Many modern dependency management tools (like npm, Maven, or Go Modules) are often a better alternative to submodules for managing software libraries.
