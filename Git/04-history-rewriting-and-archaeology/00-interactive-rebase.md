
# 00-interactive-rebase.md

- **Purpose**: To provide a comprehensive guide to `git rebase -i`, the swiss-army knife of history rewriting.
- **Estimated Difficulty**: 4/5
- **Estimated Reading Time**: 50 minutes
- **Prerequisites**: `02-advanced-local-workflows` module.

---

### What is Interactive Rebase?

A normal `git rebase <base>` replays a series of commits on top of a new base commit. An **interactive rebase** (`git rebase -i` or `--interactive`) does the same thing, but it pauses before replaying and presents you with a list of the commits it's about to move. It allows you to edit that list, giving you complete control over the history you are about to create.

You can:
- **Reorder** commits.
- **Reword** commit messages.
- **Edit** a commit (make changes to its content).
- **Squash** multiple commits into one.
- **Fixup** a commit (squash it and discard its message).
- **Drop** (delete) a commit entirely.

### The Golden Rule of Rebasing

**Do not rebase commits that you have shared with others on a public, shared branch (like `main`).**

Rebasing creates *new* commits. If you rebase a branch that others are working on, you are rewriting the shared history, which will cause massive problems for everyone else (as seen in the `03-collaboration-and-remotes` module).

Interactive rebase is primarily for cleaning up your **local, private** history before you share it with others in a Pull Request.

### How to Start an Interactive Rebase

The command is `git rebase -i <base>`. The `<base>` is the commit *before* the series of commits you want to edit.

- `git rebase -i HEAD~3`: Edit the last 3 commits.
- `git rebase -i main`: Edit all commits on your current branch that are not on `main`. (This is the most common use case for a feature branch).

When you run the command, Git opens your default text editor with a "todo" list.

```
pick 1a2b3c4 Feat: Add user model
pick 5d6e7f8 Feat: Add user controller
pick 9g8h9i0 Fix: Typo in controller

# Rebase 7k8l9m0..9g8h9i0 onto 7k8l9m0 (3 commands)
#
# Commands:
# p, pick <commit> = use commit
# r, reword <commit> = use commit, but edit the commit message
# e, edit <commit> = use commit, but stop for amending
# s, squash <commit> = use commit, but meld into previous commit
# f, fixup <commit> = like "squash", but discard this commit's log message
# x, exec <command> = run command (the rest of the line) using shell
# d, drop <commit> = remove commit
...
```

You edit the first word of each line to tell Git what to do with that commit.

### The Commands in Detail

Let's clean up the example history above.

**1. `reword` (r)**: Change a commit message.
The commit `9g8h9i0` has a message "Fix: Typo in controller". This is a poor message. We want to make it more descriptive.

*Todo List:*
```
pick 1a2b3c4 Feat: Add user model
pick 5d6e7f8 Feat: Add user controller
reword 9g8h9i0 Fix: Typo in controller
```
When you save and close, Git will replay the first two commits, then pause and open your editor again, allowing you to change the message for `9g8h9i0`.

**2. `squash` (s) and `fixup` (f)**: Combine commits.
The typo fix probably doesn't deserve its own commit. It should be part of the controller commit. We can `squash` it into the previous one.

*Todo List:*
```
pick 1a2b3c4 Feat: Add user model
pick 5d6e7f8 Feat: Add user controller
squash 9g8h9i0 Fix: Typo in controller
```
When you save, Git applies `5d6e7f8`, then applies `9g8h9i0`, and then pauses and opens an editor showing you the commit messages from *both* commits. It asks you to create a new, combined commit message.

`fixup` is the same as `squash`, but it automatically discards the second commit's message, which is perfect for small fixes.

*Todo List (better):*
```
pick 1a2b3c4 Feat: Add user model
pick 5d6e7f8 Feat: Add user controller
fixup 9g8h9i0 Fix: Typo in controller
```
This will combine the two commits and keep only the message "Feat: Add user controller".

**3. `reorder`**: Change commit order.
To reorder, you simply cut and paste the lines in the todo list. Let's say you wanted the controller to come before the model.

*Todo List:*
```
pick 5d6e7f8 Feat: Add user controller
pick 1a2b3c4 Feat: Add user model
```
Git will apply the commits in the new order. **Warning**: This can easily cause conflicts if a later commit depends on code from an earlier one.

**4. `edit` (e)**: Modify a commit's content.
You realize you forgot to add a file in the first commit.

*Todo List:*
```
edit 1a2b3c4 Feat: Add user model
pick 5d6e7f8 Feat: Add user controller
```
Git will apply commit `1a2b3c4` and then pause the rebase, dropping you back into your terminal.

```
You are currently editing a commit while rebasing branch 'my-feature' on '...'.
(use "git commit --amend" to amend the current commit)
(use "git rebase --continue" once you are satisfied with your changes)
```
You can now make any changes you want: add a file, remove a file, edit code.
```bash
$ echo "Forgot this file" > new-file.txt
$ git add new-file.txt
$ git commit --amend --no-edit # Add the staged changes to the commit
$ git rebase --continue # Continue with the rest of the rebase
```

**5. `drop` (d)**: Delete a commit.
You realize a commit was a mistake and want it gone.

*Todo List:*
```
pick 1a2b3c4 Feat: Add user model
drop 5d6e7f8 Feat: Add user controller
```
Or, you can just delete the line entirely. When you save, Git will simply not apply that commit.

### Aborting a Rebase

If you get into a mess, you can always escape with:
`git rebase --abort`
This will immediately stop the rebase and return your branch to the state it was in before you started.

### Key Takeaways

- Interactive rebase is your primary tool for cleaning up local commit history before sharing it.
- It allows you to reorder, reword, edit, squash, and drop commits.
- The most common use case is `git rebase -i main` on a feature branch to prepare it for a pull request.
- Never rebase a shared, public branch.
- Use `git rebase --abort` to escape if you get confused.

### Interview Notes

- **Question**: "Your feature branch has 15 commits, including 'fix typo', 'oops', and 'wip'. How do you clean this up before submitting a pull request that your team wants to have only a single, clean commit?"
- **Answer**: "The best tool for this is interactive rebase. I would run `git rebase -i main`. This will open an editor with a list of my 15 commits. I would leave the first commit as `pick` and change the next 14 commits to `squash` (or `fixup` if I don't care about their messages). When I save, Git will combine all 15 commits into one and then prompt me to write a new, clean commit message that summarizes the entire feature. After the rebase is complete, I'll have a single, perfect commit on my branch, ready for the PR."
