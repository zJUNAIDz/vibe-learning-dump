
# 00-the-reflog-your-safety-net.md

- **Purpose**: To introduce the reflog as Git's primary safety net, explaining what it is, what it tracks, and how to use it to recover from common mistakes.
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 40 minutes
- **Prerequisites**: `02-advanced-local-workflows/02-git-reset-demystified.md`

---

### What is the Reflog?

If the commit graph is the official history of your project, the **reflog** is your personal diary. It's a private log, local to your repository, that records almost every action you take that changes the `HEAD` pointer or any branch tip.

- It records when you switch branches (`checkout`/`switch`).
- It records when you make a commit.
- It records when you `reset` a branch.
- It records when you `rebase` or `merge`.

The reflog is your ultimate safety net. **It is almost impossible to truly lose work in Git if you know how to use the reflog.**

### How to View the Reflog

The command is simple: `git reflog`.

```bash
$ git reflog
a1b2c3d (HEAD -> main) HEAD@{0}: commit: Add new feature
e4f5g6h HEAD@{1}: reset: moving to HEAD~1
a1b2c3d (HEAD -> main) HEAD@{2}: commit: Add new feature
7k8l9m0 HEAD@{3}: checkout: moving from feature to main
...
```

The output shows:
- The commit SHA at that point in time (`a1b2c3d`).
- The reflog pointer (`HEAD@{0}`). `HEAD@{0}` is where `HEAD` is now. `HEAD@{1}` is where `HEAD` was one move ago.
- The action that was taken (`commit`, `reset`, etc.).
- A short description of the action.

### Scenario 1: Recovering a "Lost" Commit after a Hard Reset

This is the most common reflog use case. You've been working on a feature, you made a commit, and then you accidentally ran `git reset --hard HEAD~1`, seemingly deleting your work.

**1. The Mistake**
```bash
# You have a nice commit
$ git log -1
# commit a1b2c3d... My new feature

# You make a mistake
$ git reset --hard HEAD~1
# HEAD is now at e4f5g6h...
```
Your commit `a1b2c3d` is gone from the `git log`. It seems lost.

**2. Consult the Reflog**
```bash
$ git reflog
e4f5g6h HEAD@{0}: reset: moving to HEAD~1
a1b2c3d HEAD@{1}: commit: My new feature
...
```
There it is! The reflog shows that one move ago (`HEAD@{1}`), `HEAD` was pointing to commit `a1b2c3d`.

**3. Recovery**
You can now recover this commit in several ways.

- **Option A: Reset back to it.**
  ```bash
  $ git reset --hard a1b2c3d
  # Or, using the reflog pointer:
  $ git reset --hard HEAD@{1}
  ```
  Your branch is now back to the state it was in before the bad reset.

- **Option B: Create a new branch for it.** This is safer if you're not sure you want to blow away your current state.
  ```bash
  $ git switch -c recovered-feature a1b2c3d
  ```
  You now have a new branch, `recovered-feature`, that contains your "lost" work, which you can inspect and merge.

### Scenario 2: Recovering a Deleted Branch

You accidentally delete a feature branch before it was merged.

```bash
$ git branch -D my-important-feature
# Deleted branch my-important-feature (was 7k8l9m0).
```
The branch is gone. But the reflog doesn't just track `HEAD`; it tracks all branch movements.

```bash
$ git reflog
# You'll see an entry like:
# ... HEAD@{5}: commit: Last commit on my-important-feature
```
You can find the last commit that was on that branch. Once you have the SHA, you can restore the branch:

```bash
$ git switch -c my-important-feature 7k8l9m0
```
The branch is back.

### Reflog Expiration

The reflog is not kept forever. Entries expire after a certain time to save space.
- **Unreachable** entries (like the commits "lost" after a rebase) expire after **30 days** by default.
- **Reachable** entries expire after **90 days** by default.

This means you have a generous window (usually 30 days) to recover any work you thought you lost.

### Key Takeaways

- The reflog is a local, private log of all your `HEAD` and branch movements.
- It is your primary safety net for recovering from mistakes like bad resets, rebases, or deleted branches.
- `git reflog` shows you this history.
- You can use reflog pointers (e.g., `HEAD@{1}`) or the SHAs from the reflog output to restore your repository to a previous state.
- Work is not truly lost in Git until it has been unreachable for the reflog expiry period (default 30 days) and has been garbage collected.

### Recovery Drills

1.  **Hard Reset Drill**:
    - Create a new commit.
    - Run `git reset --hard HEAD~1`.
    - Use the reflog to find the SHA of the commit you just "deleted".
    - Restore your branch to that commit.
2.  **Branch Deletion Drill**:
    - Create a new branch and make a commit on it.
    - Switch back to `main` and delete the new branch using `git branch -D`.
    - Use the reflog to find the tip of the deleted branch.
    - Re-create the branch from that commit SHA.

### Interview Notes

- **Question**: "I've just run `git reset --hard` and realized I've deleted a commit I needed. Is my work gone forever?"
- **Answer**: "No, the work is not gone. Git's reflog is a safety net for exactly this situation. I would run `git reflog` to see a log of all recent movements of the `HEAD` pointer. I would find the entry corresponding to the state right before the bad reset, which will show the SHA of the commit I thought I lost. I can then use `git reset --hard <commit-sha>` to restore my branch to that state, recovering the commit. As long as the commit was made within the reflog's expiry window, which is typically 30 days for unreachable objects, the work is recoverable."
