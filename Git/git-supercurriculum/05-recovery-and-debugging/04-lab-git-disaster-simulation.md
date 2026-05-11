
# 04-lab-git-disaster-simulation.md

- **Purpose**: A hands-on lab to practice recovery skills by intentionally creating and then fixing a series of common Git disasters.
- **Estimated Difficulty**: 5/5
- **Estimated Time**: 90 minutes
- **Prerequisites**: All previous lessons in Module 05.

---

### Goal

To build confidence in recovery by simulating several "oh no" moments in a safe environment and using the appropriate tools (`reflog`, `reset`, `cherry-pick`) to fix them.

### Setup

Create a new repository for this lab.

```bash
$ mkdir disaster-lab && cd disaster-lab
$ git init
$ git switch -c main
```

### Disaster 1: The Accidental Hard Reset

**1. Create the "Good" State**
Let's create a history with a few commits.

```bash
$ echo "File 1 v1" > file1.txt && git add . && git commit -m "Add file1"
$ echo "File 2 v1" > file2.txt && git add . && git commit -m "Add file2"
$ echo "Important data" > secret.txt && git add . && git commit -m "Add secret file"
```
Your log should have three commits. The last one is the one we care about.

**2. The Disaster**
You decide you didn't want `file2.txt`, so you try to undo that commit, but you make a mistake and reset too far.

```bash
# You meant to do HEAD~1, but you typed HEAD~2
$ git reset --hard HEAD~2
```
Check your log (`git log`) and your files (`ls`). The commits for `file2.txt` and `secret.txt` are gone, and so are the files.

**3. The Recovery**

- **Action**: Use `git reflog` to find the state before the reset.
- **Command**: `git reflog`
- **Analysis**: Look for the line `reset: moving to HEAD~2`. The line directly below it (e.g., `HEAD@{1}`) will contain the SHA of the commit you were at before the reset (the "Add secret file" commit).
- **Action**: Reset your branch back to that commit.
- **Command**: `git reset --hard <sha_from_reflog>`
- **Verification**: Run `git log` and `ls`. Your commits and files should be restored.

### Disaster 2: The Mangled Interactive Rebase

**1. Create the "Good" State**
Let's create a feature branch with a messy history that we want to clean up.

```bash
$ git switch -c feature/new-login
$ echo "wip" > login.js && git add . && git commit -m "wip"
$ echo "function login() {}" > login.js && git add . && git commit -m "Add login function"
$ echo "fix typo" >> login.js && git add . && git commit -m "fixup!"
```
You have three commits you want to squash into one.

**2. The Disaster**
You start an interactive rebase, but you get confused.

```bash
$ git rebase -i main
```
In the editor, you accidentally `drop` the "Add login function" commit and save. The rebase completes, but now your `login.js` file is just `wip` and the main logic is gone. `git rebase --abort` is no longer an option.

**3. The Recovery**

- **Action**: Use `git reflog` to find the state of your branch before the rebase.
- **Command**: `git reflog`
- **Analysis**: Look for the line `rebase -i (start): checkout main`. The line just above it represents the tip of your `feature/new-login` branch before the rebase started.
- **Action**: Reset your feature branch back to that commit.
- **Command**: `git reset --hard <sha_from_reflog>`
- **Verification**: Check `git log` on your feature branch. The original three messy commits should be restored, ready for you to attempt the rebase again.

### Disaster 3: The Prematurely Deleted Branch

**1. Create the "Good" State**
You have your `feature/new-login` branch from the previous step. Let's pretend it's perfect.

**2. The Disaster**
You switch back to `main` and, in a cleanup frenzy, you delete the feature branch before it has been merged.

```bash
$ git switch main
$ git branch -D feature/new-login
```
The branch is gone.

**3. The Recovery**

- **Action**: Find the last commit that was on that branch. The `branch -D` command helpfully prints it, but let's pretend you missed it.
- **Command**: `git reflog`
- **Analysis**: The reflog tracks all `HEAD` movements. Find the line that says `checkout: moving from feature/new-login to main`. The SHA on that line is where `HEAD` was when you left the branch, which was the branch's tip.
- **Action**: Recreate the branch from that SHA.
- **Command**: `git branch feature/new-login <sha_from_reflog>`
- **Verification**: Run `git switch feature/new-login` and check the `git log`. Your work is back.

### Debrief

This lab demonstrated that even with "destructive" commands like `git reset --hard` and `git branch -D`, or a complex, failed `rebase`, no committed work was truly lost. In every case, the reflog provided a detailed history of your actions and the SHAs needed to restore your repository to a known-good state.

This should build confidence. Don't be afraid of Git's powerful commands. Be respectful of them, and know that as long as you commit your work, the reflog is your safety net.
