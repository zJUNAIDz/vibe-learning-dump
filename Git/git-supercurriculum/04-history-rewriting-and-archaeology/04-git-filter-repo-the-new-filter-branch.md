
# 04-git-filter-repo-the-new-filter-branch.md

- **Purpose**: To introduce `git filter-repo` as the modern, safe, and recommended way to perform complex history rewriting, such as removing a file from all commits.
- **Estimated Difficulty**: 5/5
- **Estimated Reading Time**: 45 minutes
- **Prerequisites**: Understanding of history rewriting dangers.

---

### The Problem: A Secret in Your History

You've just discovered that a file containing sensitive information (e.g., `credentials.env`) was accidentally committed to your repository six months ago. Even though you removed it in a later commit, it still exists in the repository's history. Anyone who clones the repository can check out the old commit and see the secret.

You need to remove this file from **every commit in your history**.

### The Old Way: `git filter-branch` (Don't Use It)

For years, the tool for this was `git filter-branch`. However, it is notoriously slow, complex, and dangerous. It has a confusing syntax and can easily corrupt your repository if used incorrectly. The official Git documentation now actively discourages its use.

### The New Way: `git filter-repo`

`git filter-repo` is a third-party tool that has become the community-standard replacement for `filter-branch`. It is:
- **Faster**: Orders of magnitude faster.
- **Safer**: It has built-in safety checks and won't let you run on a repository that isn't a fresh clone, preventing accidental data loss.
- **Simpler**: It has a much more intuitive command-line interface for common tasks.

**Installation:**
`filter-repo` is not part of Git core, so you need to install it. It's a single Python script.

```bash
# Example using pip
pip install git-filter-repo
```
Follow the official installation guide for the most up-to-date instructions.

### How to Use `filter-repo`

**The Cardinal Rule**: `filter-repo` rewrites your entire history. This is an extremely destructive operation. You should **only** run it on a fresh, mirror clone of your repository.

**Workflow for Removing a File:**

**1. Get a Fresh Clone**
Clone your repository using the `--mirror` flag. This creates a bare repository that contains all the Git data but no working directory. This is your backup.

```bash
$ git clone --mirror https://github.com/my-org/my-repo.git
$ cd my-repo.git
```

**2. Run `filter-repo`**
Now, run the command to remove the sensitive file.

```bash
$ git filter-repo --path credentials.env --invert-paths
```
- `--path credentials.env`: Specifies the file you want to operate on.
- `--invert-paths`: This is the key. It means "remove everything that matches the path filter."

`filter-repo` will go through every commit in your history. For each commit, it will remove the `credentials.env` file and create a new, rewritten commit. The result is a new history that looks as if that file never existed.

**3. Inspect the New History**
You can check that the file is gone.

```bash
# You can't run normal git log on a bare repo easily, but you can check objects.
# A better way is to clone from your filtered bare repo and inspect that.
$ cd ..
$ git clone my-repo.git my-repo-clean
$ cd my-repo-clean
$ git log --all --full-history -- "**/credentials.env"
# This log should be empty.
```

**4. Push the New History**
This is the most critical and dangerous step. You are about to replace the entire history on your remote server. **This will invalidate all open pull requests, forks, and clones of your repository.** You must coordinate this with your entire team.

```bash
# Go back to the filtered bare repo
$ cd ../my-repo.git

# Push the new history to the remote. You must force it.
$ git push origin --force --all
$ git push origin --force --tags
```

**5. Team Communication**
Every single person on your team must now delete their old local clone and re-clone the repository from the remote to get the new, clean history. If they try to pull or push from their old clone, they will re-introduce the old history or cause massive merge conflicts.

### Other Powerful Uses of `filter-repo`

- **Renaming a user**:
  `git filter-repo --name-map <(echo "old-name=>new-name") --email-map <(echo "old@email.com=>new@email.com")`

- **Extracting a subdirectory into its own repository**:
  `git filter-repo --path-rename path/to/folder/:`
  This makes the `folder` the new root of the repository.

- **Stripping large files**:
  `git filter-repo --strip-blobs-bigger-than 10M`

### Key Takeaways

- `git filter-branch` is deprecated. Use `git filter-repo`.
- `filter-repo` is the tool for completely removing a file from all of history.
- This is a highly destructive operation that rewrites all commit SHAs.
- **Always** run it on a fresh, mirror clone.
- You **must** coordinate with your entire team, as everyone will need to delete their old clones and re-clone the repository after you force-push the new history.
- This is a "break-glass" procedure for security incidents or major repository restructuring.

### Interview Notes

- **Question**: "I've accidentally committed a password to a public repository. I've since removed it and pushed a new commit, but I know it's still in the history. How do I fix this?"
- **Answer**: "First, and most importantly, you must immediately treat that password as compromised and rotate it. The secret is public. After that, to clean the repository, the correct tool is `git filter-repo`. The process is: 1. Announce a maintenance window to the team, as this will require everyone to re-clone. 2. Create a fresh `--mirror` clone of the repository. 3. Use `git filter-repo --path path/to/secret-file --invert-paths` to remove the file from every commit in history. 4. After verifying the file is gone, force-push the new history to all branches and tags on the remote. 5. Instruct all developers to delete their local clones and clone a fresh copy. Simply removing it from history is not enough; the secret must be considered compromised from the moment it was pushed."
