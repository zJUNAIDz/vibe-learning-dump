
# 02-recovering-deleted-branches-and-commits.md

- **Purpose**: To provide a clear playbook for recovering branches or commits that have been accidentally deleted.
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 30 minutes
- **Prerequisites**: `00-the-reflog-your-safety-net.md`

---

### The Fear of Deletion

Losing work is a developer's worst nightmare. Fortunately, Git's data model makes it very difficult to truly lose anything that has been committed. As long as a commit object exists in the database and is reachable from the reflog, it can be recovered.

### Scenario 1: Recovering a Deleted Branch

You're cleaning up your local branches and accidentally delete a feature branch that wasn't fully merged yet.

```bash
$ git branch -d feature/user-profile # This would fail if not merged.
$ git branch -D feature/user-profile # You force-deleted it!
Deleted branch feature/user-profile (was a1b2c3d).
```

The branch pointer is gone, but the commits are still there.

**The Recovery Playbook:**

1.  **Don't Panic.** The work is not gone.
2.  **Consult the Reflog.** The reflog tracks not just `HEAD`, but the tips of all branches.
    ```bash
    $ git reflog
    ...
    e4f5g6h HEAD@{2}: checkout: moving from feature/user-profile to main
    a1b2c3d HEAD@{3}: commit: Feat: Add user avatar
    ...
    ```
    The reflog shows the last commit made on that branch (`a1b2c3d`). The message from the `git branch -D` command also helpfully told you the SHA.

3.  **Recreate the Branch.** Once you have the SHA of the commit that was at the tip of the branch, you can recreate the branch instantly.
    ```bash
    $ git branch feature/user-profile a1b2c3d
    ```
    Or, if you want to check it out at the same time:
    ```bash
    $ git switch -c feature/user-profile a1b2c3d
    ```
The branch is now fully restored.

### Scenario 2: Recovering a Deleted Commit (via `reset`)

You have a single commit you want to get rid of, so you use `reset`.

```bash
$ git log --oneline
a1b2c3d (HEAD -> main) The commit I want to delete
e4f5g6h The commit before it

$ git reset --hard e4f5g6h
```
Commit `a1b2c3d` is no longer in the log for the `main` branch.

**The Recovery Playbook:**

1.  **Consult the Reflog.**
    ```bash
    $ git reflog
    e4f5g6h HEAD@{0}: reset: moving to e4f5g6h
    a1b2c3d HEAD@{1}: commit: The commit I want to delete
    ...
    ```
    The reflog clearly shows the "lost" commit at `HEAD@{1}`.

2.  **Decide How to Recover.** You have options:
    - **Restore the branch completely:**
      `$ git reset --hard a1b2c3d`

    - **Cherry-pick the commit:** Maybe you want the *changes* from that commit, but applied on top of some other work.
      ```bash
      # You are on a different branch now...
      $ git cherry-pick a1b2c3d
      ```

    - **Create a new branch to inspect it:**
      `$ git branch temporary-recovery-branch a1b2c3d`

The key is that the reflog gives you the SHA, and the SHA is your handle to the commit object. Once you have the handle, you can do anything you want with it.

### What About Uncommitted Work?

The reflog and Git's object database can only save you if you have **committed** your work.

**If you use `git reset --hard` when you have uncommitted changes in your working directory or index, that work is gone forever.** It never made it into the object database, so Git has no way to recover it.

This is why frequent, small commits are a good habit. A commit is a safety net. The more frequently you commit, the more granular your recovery options are.

### `git fsck`: Finding Dangling Objects

What happens to commits after they become "unreachable" (e.g., after a reset or a rebase)? They are called "dangling" objects. They are still in the database but nothing points to them.

The `git fsck` (filesystem check) command can be used to find these objects.

```bash
$ git fsck --lost-found
Checking object directories: 100% (256/256), done.
dangling commit a1b2c3d...
dangling commit 7k8l9m0...
```
This can be another way to find lost commits if the reflog is somehow unclear, but it's a much lower-level tool. It just tells you the objects exist, with no context about what branch they were on or when they were created. The reflog is almost always the better tool for recovery.

### Key Takeaways

- It is very difficult to lose committed work in Git.
- The `reflog` is the primary tool for finding the SHAs of deleted commits or branch tips.
- Once you have the SHA, you can use `git branch`, `git switch -c`, `git reset`, or `git cherry-pick` to recover the work.
- **Uncommitted work is not protected.** `git reset --hard` will permanently delete unstaged changes. Commit often.

### Recovery Drill

1.  Create a new branch `test-delete`.
2.  Make two new commits on this branch.
3.  Switch back to `main`.
4.  Delete the branch with `git branch -D test-delete`.
5.  Use `git reflog` to find the SHA of the second commit you made on that branch.
6.  Restore the branch to its full state using that SHA.
7.  Now, on the restored branch, do `git reset --hard HEAD~2` to delete both commits.
8.  Use `git reflog` again to find the SHAs of the two commits you just removed.
9.  Use `git cherry-pick` to re-apply them to your branch.
