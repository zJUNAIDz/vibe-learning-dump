
# 02-git-reset-demystified.md

- **Purpose**: To provide a crystal-clear, systems-level explanation of `git reset` by explaining its effect on the three trees.
- **Estimated Difficulty**: 4/5
- **Estimated Reading Time**: 40 minutes
- **Prerequisites**: `01-the-many-faces-of-git-checkout.md`

---

### The Most Feared Command

`git reset` is arguably the most powerful and feared command in Git. Its reputation comes from its ability to "rewrite history" and "lose work." But like all Git commands, it's perfectly safe and predictable once you understand its effect on the three trees: **HEAD**, the **Index**, and the **Working Directory**.

`git reset` fundamentally does one thing: **it resets the current HEAD to a specified state.** The different "modes" (`--soft`, `--mixed`, `--hard`) determine how it affects the Index and the Working Directory.

Let's assume `HEAD` is currently at commit `C`. We run `git reset B`.

```mermaid
graph LR
    A --> B --> C -- "HEAD, master"
```

### `git reset --soft <commit>`

This is the safest mode. It only moves the `HEAD` pointer.

- **1. Moves HEAD**: The branch `HEAD` points to (e.g., `master`) is moved to point to the target commit (`B`).
- **2. Index**: Unchanged. It still matches the state of the original commit (`C`).
- **3. Working Directory**: Unchanged.

```mermaid
graph LR
    A --> B -- "HEAD, master"
    B --> C
```
**After `git reset --soft B`:**
- `HEAD` is now at `B`.
- The Index still contains the changes from `C`.
- The Working Directory is untouched.

**Result**: `git status` will show all the changes from commit `C` as "Changes to be committed" (staged).

**Use Case**: "I want to undo my last commit, but keep all the changes staged. I'm going to re-commit them differently, perhaps as multiple commits or with a better message." This is great for squashing or editing the last commit.

### `git reset --mixed <commit>` (The Default)

This is the default mode if you don't specify one. It moves `HEAD` and updates the `Index`.

- **1. Moves HEAD**: The `master` branch pointer is moved to `B`.
- **2. Index**: Is updated to match the state of the new `HEAD` (`B`).
- **3. Working Directory**: Unchanged.

**Result**: `git status` will show all the changes from commit `C` as "Changes not staged for commit" (unstaged). The changes are still safely in your working directory.

**Use Case**: "I want to undo my last commit and all the staging I did for it. I want to re-evaluate all the changes and decide what to stage and commit from scratch."

### `git reset --hard <commit>`

This is the most "dangerous" mode because it's the only one that can destroy uncommitted work. It moves `HEAD` and updates both the `Index` and the `Working Directory`.

- **1. Moves HEAD**: The `master` branch pointer is moved to `B`.
- **2. Index**: Is updated to match the state of the new `HEAD` (`B`).
- **3. Working Directory**: Is updated to match the state of the new `HEAD` (`B`).

**Result**: All changes from commit `C`, plus any unstaged changes you had in your working directory, are **gone**. Your entire local state (all three trees) is reset to exactly match commit `B`.

**Use Case**: "I want to completely throw away my last commit and any other local changes I've made. I want my repository to be in the exact state it was at commit `B`." This is a "nuke it from orbit" command.

**Is the work really lost?** The commit object `C` is now "unreachable," but it still exists in the object database for a while. You can find its SHA in the `reflog` and get it back. However, any *uncommitted* changes in your working directory are truly gone forever.

### Summary Table

| Mode      | Moves `HEAD`? | Updates `Index`? | Updates `Working Directory`? | Use Case                               |
| --------- | :-----------: | :--------------: | :--------------------------: | -------------------------------------- |
| `--soft`  |      Yes      |        No        |              No              | Undo commit, keep changes staged.      |
| `--mixed` |      Yes      |       Yes        |              No              | Undo commit and staging, keep changes. |
| `--hard`  |      Yes      |       Yes        |             Yes              | Nuke everything since `<commit>`.      |

### `git reset` on a Path

You can also run `git reset` on a specific file path, e.g., `git reset -- my-file.txt`. This is a legacy command. In this mode, it does **not** move `HEAD`. It's a way to unstage a file.

`git reset HEAD my-file.txt` is equivalent to the modern `git restore --staged my-file.txt`. It copies the file's state from `HEAD` to the `Index`. It's highly recommended to use `restore` for this, as it's much clearer.

### Key Takeaways

- `git reset` moves the `HEAD` pointer.
- The mode (`--soft`, `--mixed`, `--hard`) determines how far the "reset wave" travels through the three trees.
- `--soft`: Resets `HEAD` only.
- `--mixed`: Resets `HEAD` and `Index`.
- `--hard`: Resets `HEAD`, `Index`, and `Working Directory`.
- Never use `--hard` if you have uncommitted changes in your working directory that you want to keep.
- Use `git reflog` to recover from an accidental `reset`.

### Interview Notes

- **Question**: "Explain the difference between `git reset --soft`, `git reset --mixed`, and `git reset --hard`."
- **Answer**: "All three modes of `git reset` move the `HEAD` pointer to a specified commit. The difference is how they affect the other two trees: the index and the working directory. `--soft` only moves `HEAD`, leaving the index and working directory untouched. This has the effect of 'un-committing' changes but leaving them staged. `--mixed`, the default, moves `HEAD` and also resets the index to match `HEAD`, leaving the changes in the working directory as unstaged. `--hard` moves `HEAD` and resets both the index and the working directory to match `HEAD`, discarding all changes, both staged and unstaged, that were made after the target commit."
