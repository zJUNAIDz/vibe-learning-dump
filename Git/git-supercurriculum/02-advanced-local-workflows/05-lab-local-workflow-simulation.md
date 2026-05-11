
# 05-lab-local-workflow-simulation.md

- **Purpose**: A hands-on lab simulating a complex local workflow involving feature work, an interruption for a hotfix, and using advanced local commands to manage the state.
- **Estimated Difficulty**: 4/5
- **Estimated Time**: 60 minutes
- **Prerequisites**: All previous lessons in Module 02.

---

### Goal

To simulate a realistic "day in the life" of a developer, using `switch`, `restore`, `stash`, `reset`, and `worktree` to efficiently manage multiple streams of work.

### Scenario

You are a developer working on a new feature, "user-profiles". You've done some work, but it's not finished. Suddenly, a critical production bug is reported that needs your immediate attention. While you're fixing that, a colleague asks for a quick review of a small change.

### Setup

```bash
$ mkdir workflow-lab && cd workflow-lab
$ git init
$ git switch -c main

# Create the initial state of the project
$ echo "Homepage v1" > index.html
$ mkdir js
$ echo "console.log('App loaded');" > js/app.js
$ git add .
$ git commit -m "Initial commit"
```

### Part 1: The Feature Work

You start working on the "user-profiles" feature.

1.  **Create a feature branch.**
    ```bash
    $ git switch -c feature/user-profiles
    ```
2.  **Do some work.** You add a new file and modify an existing one.
    ```bash
    $ echo "User Profile Page" > profile.html
    $ echo "console.log('App loaded with profile feature');" > js/app.js
    ```
3.  **Stage some of the work.** You're happy with the new file, but the change to `app.js` is just for debugging.
    ```bash
    $ git add profile.html
    ```
4.  **Check your status.**
    ```bash
    $ git status
    # You should see profile.html as staged, and js/app.js as unstaged.
    ```

**At this point, the critical bug report comes in. You must stop your work immediately.**

### Part 2: The Interruption (Hotfix) - Stash Method

First, we'll handle the interruption using the `stash` workflow.

1.  **Stash your work.** You need to save everything: the staged `profile.html`, the unstaged `js/app.js`, and you should leave a message.
    ```bash
    $ git stash push -m "WIP: user-profiles, profile page added, app.js has debug logs"
    ```
2.  **Verify your workspace is clean.**
    ```bash
    $ git status # Should be clean
    ```
3.  **Switch back to `main` to start the hotfix.**
    ```bash
    $ git switch main
    ```
4.  **Create a hotfix branch.**
    ```bash
    $ git switch -c hotfix/fix-homepage-title
    ```
5.  **Fix the bug.** The bug is a typo in the homepage.
    ```bash
    $ echo "Homepage v1.0.1" > index.html
    ```
6.  **Commit the fix.**
    ```bash
    $ git add index.html
    $ git commit -m "Fix: Correct homepage title typo"
    ```
7.  **Merge the fix back into `main`.** (In a real workflow, this would be a PR).
    ```bash
    $ git switch main
    $ git merge --no-ff hotfix/fix-homepage-title
    ```

### Part 3: Returning to Feature Work

Now that the hotfix is "deployed," you can return to your feature.

1.  **Switch back to your feature branch.**
    ```bash
    $ git switch feature/user-profiles
    ```
2.  **Bring your feature branch up to date.** It's good practice to rebase your feature branches on the latest `main`.
    ```bash
    $ git rebase main
    ```
3.  **Get your stashed work back.**
    ```bash
    $ git stash pop
    # It should apply cleanly.
    ```
4.  **Check your status.**
    ```bash
    $ git status
    # You should be back to exactly where you were at the end of Part 1.
    # profile.html is staged, js/app.js is unstaged.
    ```

### Part 4: The Code Review - Worktree Method

Let's reset and try the interruption scenario again, but this time using `git worktree`.

1.  **Reset to the start of the interruption.**
    ```bash
    # First, clean up the stash and the hotfix commits
    $ git reset --hard HEAD~2 # Go back before the merge and hotfix commits
    $ git switch feature/user-profiles
    $ git stash clear
    $ git reset --hard <commit_sha_from_end_of_part_1> # You may need to find this in reflog
    # Or, just recreate the state from Part 1 manually.
    ```
2.  **You are on `feature/user-profiles` with staged and unstaged work.** A colleague asks for a code review. Instead of stashing, create a worktree.
    ```bash
    $ git worktree add ../pr-review-temp feature/colleague-branch
    # Assuming 'feature/colleague-branch' is the branch they want you to review.
    ```
3.  **Perform the review.**
    ```bash
    $ cd ../pr-review-temp
    # You are in a totally clean directory on the other branch.
    # You can run tests, make comments, etc.
    $ cd ../workflow-lab
    # Your original directory is completely untouched.
    ```
4.  **Clean up the review worktree.**
    ```bash
    $ git worktree remove pr-review-temp
    ```

### Debrief and Comparison

- **Stash Workflow**:
    - **Pros**: Quick for simple interruptions. Keeps everything in one directory.
    - **Cons**: Can lead to stash-stacking and confusion. `stash pop` can cause conflicts. You lose the context of your running processes.
- **Worktree Workflow**:
    - **Pros**: Completely isolated environments. No need to stash. You can have multiple branches "active" at once. Preserves dev server state, `node_modules`, etc.
    - **Cons**: Creates extra directories on your filesystem. Requires a bit more setup.

For a quick context switch, `stash` is fine. For anything that will take more than a few minutes (like a code review or a complex hotfix), `git worktree` is a superior, safer, and more powerful workflow.
