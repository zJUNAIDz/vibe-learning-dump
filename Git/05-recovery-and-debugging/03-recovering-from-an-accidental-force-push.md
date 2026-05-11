
# 03-recovering-from-an-accidental-force-push.md

- **Purpose**: To provide a recovery plan for the emergency scenario where someone has accidentally force-pushed to a shared branch like `main`.
- **Estimated Difficulty**: 5/5
- **Estimated Reading Time**: 40 minutes
- **Prerequisites**: `00-the-reflog-your-safety-net.md`, `03-collaboration-and-remotes/01-the-push-and-pull-dance.md`

---

### The Ultimate "Oh No" Moment

Someone on your team has accidentally run `git push --force` on `main`. The remote history has been overwritten, and several commits from other team members appear to be gone. This is a "stop everything and fix it now" emergency.

The good news is that as long as *someone* has the correct commits, they are recoverable. The remote repository's reflog might even save you.

### The Recovery Plan

**1. LOCK THE BRANCH**
Immediately go to your repository hosting platform (GitHub, GitLab) and protect the `main` branch. Configure it so that force-pushes are disabled and that all merges require a pull request. This prevents the situation from getting worse while you fix it.

**2. FIND THE LOST COMMITS**
You need to find the commit that `main` *should* be pointing to. There are a few places to look:

- **A Teammate's Local Machine**: The best-case scenario. Someone on the team has not yet run `git pull`. Their local `main` branch still points to the correct, pre-force-push history. They are the source of truth.

- **The Remote's Reflog (GitHub/GitLab)**: Many Git hosting providers keep a reflog on the server side, though it may not be directly accessible. However, the "Activity" or "Push" events feed for the repository will almost certainly show the force push and, crucially, the commit SHA that `main` pointed to *before* it was overwritten. This is your golden ticket.

- **CI/CD Server**: Your build server (e.g., Jenkins, GitHub Actions) likely checked out the correct commit to run tests. The logs for the last successful build on `main` will contain the full SHA of the commit you need to restore.

**3. THE RECOVERY (Performed by one designated person)**

Let's assume you've found the correct SHA, let's call it `a1b2c3d`.

**Step 3a: Restore the branch locally**
The designated person fixing the issue needs to get their local `main` into the correct state.

```bash
# Ensure you have the latest objects from the remote
$ git fetch origin

# Reset your local main branch to the correct, pre-force-push commit
$ git switch main
$ git reset --hard a1b2c3d
```
Your local `main` branch is now correct.

**Step 3b: Push the corrected branch to the remote**
Now you need to update the remote to match your corrected local branch. Because the remote has a divergent history (from the bad force push), you will need to force-push again. This is the one time it is acceptable, because you are force-pushing to fix a previous force-push.

```bash
# You are on main, which now points to the correct commit.
$ git push --force-with-lease origin main
```
Using `--force-with-lease` is still best practice, as it ensures no one else has pushed to `main` in the few minutes you've been working on the fix.

**4. COMMUNICATE TO THE TEAM**
The `main` branch on the remote is now fixed. You must now instruct every single person on the team on how to fix their local repositories.

Send out a clear message:
> "The `main` branch has been restored after an accidental force-push. To get your local repository back in sync, please run the following commands for your local `main` branch:
>
> `git switch main`
> `git fetch origin`
> `git reset --hard origin/main`
>
> **DO NOT PULL OR MERGE `main`**. You must reset it."

If anyone merges instead of resetting, they will re-introduce the messy, divergent history, and you'll be back at square one.

### What if the Bad Push Also Contained New Work?

This is a more complex scenario. The person who force-pushed may have done so to publish their own new commits, but they did it on top of an old version of `main`.

**Recovery:**
1.  Follow steps 1-4 above to restore `main` to the state *before* the bad push.
2.  The person who made the bad push now needs to recover their own work. Their commits are in their local reflog.
3.  They should check out a new branch (`git switch -c my-recovered-work`) at the point before their bad push.
4.  They can then `cherry-pick` their new commits from their reflog onto this new branch.
5.  This new branch can then be rebased on top of the now-fixed `main` branch.
6.  Finally, they can open a proper Pull Request to merge their work.

### Key Takeaways

- **Prevention is the best cure**: Protect your shared branches ( `main`, `develop`) from force-pushes in your remote repository's settings.
- **Act fast**: Lock the branch and communicate immediately.
- **Find the true `HEAD`**: The most important step is finding the SHA of the last good commit from a teammate's repo, CI logs, or the remote's activity feed.
- **Reset, don't merge**: The recovery process involves one person force-pushing the fix, and everyone else using `git reset --hard` to update their local copies.
- Even this "ultimate" disaster is recoverable if you stay calm and understand the mechanics.

### Recovery Drill

This is hard to simulate alone, but you can approximate it:
1.  Have two clones of a bare repository.
2.  In clone A, make a commit on `main` and push it. (`Commit A`)
3.  In clone B, run `git pull`.
4.  In clone A, make another commit. (`Commit B`)
5.  In clone B, run `git reset --hard HEAD~1` (so you are back at `Commit A`). Then make a new, different commit. (`Commit C`).
6.  In clone B, run `git push --force origin main`. The remote `main` now points to `Commit C`. `Commit B` is "lost".
7.  Now, in clone A, your `origin/main` is out of date. Run `git fetch`. Your local `main` points to `Commit B`, but `origin/main` now points to `Commit C`. You are in the position of the person with the "correct" history.
8.  Practice the recovery: `git push --force-with-lease origin main` from clone A to restore the remote.
9.  Then, go to clone B and practice the team recovery: `git reset --hard origin/main`.
