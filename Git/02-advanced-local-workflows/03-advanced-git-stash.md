
# 03-advanced-git-stash.md

- **Purpose**: To explore advanced use cases and internals of `git stash`, moving beyond simple "save and restore."
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 30 minutes
- **Prerequisites**: `02-git-reset-demystified.md`

---

### `git stash` is Not a Clipboard

Many developers think of `git stash` as a simple clipboard for their changes. While this is a useful starting analogy, it hides the power and mechanics of what `stash` actually does.

**A stash is a commit object.** Or more accurately, a set of commit objects.

When you run `git stash`, Git does the following:
1.  It creates a **commit object** to store the state of your `Index` (staged changes).
2.  It creates a second **commit object** to store the state of your `Working Directory` (unstaged changes).
3.  It creates a third **commit object** that acts as a "merge commit," with the two commits above as its parents, and the current `HEAD` as its *first* parent.
4.  It stores the SHA of this third commit in a special ref located at `.git/refs/stash`.
5.  It runs `git reset --hard HEAD` to clean your working directory.

**Diagram: The Stash Commit Structure**
```mermaid
graph TD
    subgraph "Commit Graph"
        HEAD -- "Current Commit"
    end

    subgraph "Stash Object (a special commit)"
        S["Stash Commit"]
        W["Working Dir Commit"]
        I["Index Commit"]

        S -- "parent 1" --> HEAD
        S -- "parent 2" --> W
        S -- "parent 3 (optional)" --> I
    end
```
This is why stashing is so robust. It's not some temporary patch file; it's using Git's own powerful commit and object model.

### Managing Multiple Stashes

The stash is not a single slot; it's a stack.

- `git stash`: Pushes a new stash onto the stack.
- `git stash list`: Shows all stashes on the stack (`stash@{0}`, `stash@{1}`, etc.).
- `git stash apply`: Applies the top stash (`stash@{0}`) but leaves it on the stack.
- `git stash pop`: Applies the top stash and then removes it from the stack if successful.
- `git stash apply stash@{2}`: Applies a specific stash from the stack, not just the top one.
- `git stash drop stash@{2}`: Deletes a specific stash from the stack.
- `git stash clear`: Deletes all stashes.

### Advanced Stashing Techniques

**1. Stashing with a Message**
Stashes can quickly become cryptic. Always leave a message.

```bash
$ git stash push -m "WIP: Refactoring user auth, not yet compiling"
```
Now `git stash list` will show this message, making it much easier to find the right stash later.

**2. Stashing Untracked and Ignored Files**
By default, `git stash` only stashes modified tracked files.
- `git stash -u` (or `--include-untracked`): Stashes untracked files as well.
- `git stash -a` (or `--all`): Stashes untracked *and* ignored files.

**Use Case**: You need to switch branches, but you've generated some build artifacts or logs that you want to keep with your current changes. `git stash -a` will pack everything up neatly.

**3. Stashing Only a Portion of Your Changes**
Just like `git add --patch`, you can stash interactively.

```bash
$ git stash -p
```
Git will walk you through each "hunk" of your changes and ask if you want to stash it. This is incredibly powerful for separating work.

**Scenario**: You've fixed a bug and started a new feature in the same file. A hotfix request comes in.
1.  `git stash -p` and stash only the new feature work.
2.  Commit and push the bug fix.
3.  `git stash pop` to get your feature work back.

### `git stash branch`: The Safest Way to Work with Stashes

What if you `pop` a stash and it has massive conflicts? Or what if you're not sure you want to merge it into your current branch?

`git stash branch <branch-name> <stash>` creates a new branch based on the commit where you *created* the stash, and then applies the stash to it.

```bash
# You are on 'main', but you have an old stash from a 'feature' branch
$ git stash branch temp-feature-work stash@{1}
```
This command will:
1.  Check out the commit that `feature` was on when you made `stash@{1}`.
2.  Create a new branch called `temp-feature-work`.
3.  Apply the stashed changes.
4.  Drop `stash@{1}`.

You now have a clean branch with your stashed work, ready to be reviewed, rebased, or merged without polluting your current branch. This is the single best way to recover old or complex stashes.

### Key Takeaways

- A stash is a set of real commit objects.
- The stash is a stack; you can manage multiple stashes.
- Use `git stash push -m "message"` to keep your stashes organized.
- Use `git stash -p` for partial stashing.
- Use `git stash branch` to safely recover and inspect complex or old stashes.

### Interview Notes

- **Question**: "I ran `git stash pop` and got a huge merge conflict. What's a safe way to handle this?"
- **Answer**: "First, I would abort the messy pop with `git reset --hard`. The stash is not dropped if the apply fails, so it's still safe. A much safer way to deal with a complex stash is to not apply it to your current work at all. Instead, use `git stash branch <new-branch-name> <stash>`. This creates a new branch at the exact commit where the stash was originally made and applies the changes there. This isolates the stashed work in a clean environment, where you can resolve any issues, turn it into proper commits, and then decide how to integrate it back into your main branch, probably via a rebase or merge."
