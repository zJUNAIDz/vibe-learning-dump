
# 02-merge-conflicts-a-deep-dive.md

- **Purpose**: To explain what merge conflicts are at a systems level and provide a structured, calm approach to resolving them.
- **Estimated Difficulty**: 4/5
- **Estimated Reading Time**: 50 minutes
- **Prerequisites**: `01-the-push-and-pull-dance.md`

---

### What is a Merge Conflict?

A merge conflict is not an error. It is Git pausing and asking for human intervention. It occurs when Git is trying to combine two divergent histories and finds changes in the same lines of the same file that it cannot automatically resolve.

To understand this, we need to understand Git's merge strategy. When you `git merge feature`, Git performs a "three-way merge." It looks at three commits:
1.  **The tip of your current branch (`main`)**.
2.  **The tip of the branch you're merging in (`feature`)**.
3.  **The common ancestor** of these two branches (the "merge base").

Git then compares the changes.
- If a change was made in `feature` but not in `main`, Git takes the `feature` version.
- If a change was made in `main` but not in `feature`, Git keeps the `main` version.
- **If the *same lines* were changed in both `feature` and `main` relative to the common ancestor, Git cannot make a decision. This is a merge conflict.**

**Diagram: The Three-Way Merge**
```mermaid
graph TD
    subgraph "Commit Graph"
        A["Merge Base"] --> B["main"]
        A --> C["feature"]
    end

    subgraph "File States"
        F_A["File at A"]
        F_B["File at B"]
        F_C["File at C"]
    end

    GitCompares["Git Compares"]
    GitCompares -- "diff A B" --> ChangesInMain
    GitCompares -- "diff A C" --> ChangesInFeature

    subgraph "Conflict"
        Conflict["Conflict if changes overlap"]
    end

    ChangesInMain --> Conflict
    ChangesInFeature --> Conflict
```

### Anatomy of a Conflict Marker

When a conflict occurs, Git stops and places conflict markers in the affected file(s).

```
<<<<<<< HEAD
This is the version of the line from your current branch (HEAD).
=======
This is the version of the line from the branch you are merging in.
>>>>>>> feature
```

- `<<<<<<< HEAD`: The start of the conflicting block from your current branch.
- `=======`: The separator between the two conflicting versions.
- `>>>>>>> feature`: The end of the conflicting block, indicating the name of the branch the other version came from.

Your job is to edit this block, removing the markers and leaving the file in its final, correct state. This might mean keeping your version, their version, a combination of both, or something entirely new.

### A Structured Approach to Conflict Resolution

Panic is the enemy of a good merge. Follow a calm, structured process.

**1. Assess the Situation**
First, use `git status` to understand the scope of the conflict.

```bash
$ git status
On branch main
You have unmerged paths.
  (fix conflicts and run "git commit")
  (use "git merge --abort" to abort the merge)

Unmerged paths:
  (use "git add <file>..." to mark resolution)
        both modified:   src/api.js

no changes added to commit (use "git add" and/or "git commit -a")
```
This tells you exactly which files have conflicts.

**2. Resolve Each File**
Open each conflicted file (`src/api.js` in this case).
- Look for the `<<<<<<<` markers.
- Read the code in both sections.
- **Decide what the correct final version should be.** This is the most important step and requires understanding the *intent* behind both sets of changes. You may need to talk to the other developer.
- Edit the file to make it correct, removing the conflict markers.

**Example Resolution:**
*Before:*
```javascript
<<<<<<< HEAD
const port = process.env.PORT || 3000;
=======
const port = process.env.PORT || 5000;
>>>>>>> feature
```
*After deciding 5000 is the correct default:*
```javascript
const port = process.env.PORT || 5000;
```

**3. Stage the Resolved File**
Once you have saved the file, you must tell Git that you have resolved the conflict. You do this with `git add`.

```bash
$ git add src/api.js
```
This does **not** create a new blob. It moves the file from the "Unmerged paths" section of the index to the "Changes to be committed" section. The index has special slots for conflicts, and `git add` signals that the conflict in slot 0 is resolved.

**4. Finalize the Merge**
After you have resolved and `add`ed all conflicted files, `git status` will look like this:

```bash
$ git status
On branch main
All conflicts fixed but you are still merging.
  (use "git commit" to conclude merge)

Changes to be committed:
        modified:   src/api.js
```
The merge is paused, waiting for you to create the merge commit.

```bash
$ git commit
# Git will open an editor with a pre-populated merge commit message.
# You can leave it as is or add more detail. Save and close.
```
The merge is now complete.

### Tools to Help

- **Merge Tools**: You can configure a graphical merge tool (like `vscode`, `p4merge`, `kdiff3`) to help with resolution.
  - `git config --global merge.tool vscode`
  - `git mergetool` will then open the conflicted files in your chosen tool, often with a three-way or four-way view that is much easier to understand.

- **`--diff3` Conflict Style**: This style shows the original version from the merge base, which can be invaluable for understanding *why* the conflict happened.
  ```bash
  $ git config --global merge.conflictstyle diff3
  ```
  The conflict marker will now look like this:
  ```
  <<<<<<< HEAD
  My version
  ||||||| merged common ancestors
  Original version
  =======
  Their version
  >>>>>>> feature
  ```

### Aborting a Merge

If you get into a mess and want to start over, you can always abort the merge process.

```bash
$ git merge --abort
```
This will reset your repository back to the state it was in right before you ran `git merge`.

### Key Takeaways

- A merge conflict is Git asking for help when changes overlap.
- It's based on a three-way comparison between your branch, their branch, and the common ancestor.
- The resolution process is: **Edit -> Add -> Commit**.
- `git add` is the command that marks a conflict as resolved.
- Use `git merge --abort` to escape a messy merge.
- Configure `conflictstyle = diff3` to get more context on *why* a conflict occurred.

### Collaboration Pitfalls

- **Resolving conflicts you don't understand**: Never just blindly accept "mine" or "theirs." If you don't understand the other person's change, you risk re-introducing bugs. Go talk to them.
- **Forgetting to `git add`**: A common mistake is to edit the files but forget to `git add` them before committing. This can lead to the conflict markers themselves being committed to the repository, which is a major headache.
