
# 02-finding-fault-with-git-blame-and-git-log.md

- **Purpose**: To teach advanced techniques for historical investigation using `git log` and `git blame`.
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 45 minutes
- **Prerequisites**: Basic `git log` usage.

---

### `git log`: More Than Just a List

You already know `git log` shows you a list of commits. But its default output is verbose and often not what you need. Mastering `git log`'s formatting and filtering options is a superpower for understanding a project's history.

### Advanced Formatting

- `git log --oneline`: The most useful format. Shows the SHA and the commit title on a single line.
- `git log --graph --oneline --decorate`: Shows the commit history as a graph, which is essential for visualizing branches and merges. `--decorate` adds branch and tag names. **Alias this immediately.**
  - `git config --global alias.lol "log --graph --oneline --decorate"` -> `git lol`

- **Custom Formatting**: You can design your own format with `--pretty=format:"..."`.
  - `%h`: abbreviated SHA
  - `%an`: author name
  - `%ar`: author date, relative (e.g., "2 weeks ago")
  - `%s`: subject (commit title)
  - `%d`: refs (branch/tag names)

  **Example: A beautiful custom log**
  ```bash
  $ git log --pretty=format:"%C(yellow)%h %C(reset)- %C(cyan)%an %C(reset)(%C(red)%ar%C(reset))%C(auto)%d %n   %s"
  # Alias this!
  # git config --global alias.lg "log --pretty=format:'...'"
  ```

### Advanced Filtering

You don't always want to see the whole history.

- **By Author**:
  `git log --author="Alice"`

- **By Date**:
  `git log --since="2 weeks ago"`
  `git log --before="2023-01-01"`

- **By Message Content**:
  `git log --grep="Fix: login bug"`

- **By File or Directory**:
  `git log -- path/to/file.js`
  This shows only commits that touched that specific file.

- **By Content (The "Pickaxe")**:
  This is one of `log`'s most powerful features. The `-S` option (the "pickaxe") finds commits where the *number of occurrences* of a string changed. This is how you find when a specific function call or variable was introduced or removed.

  **Scenario**: When was the `getUserById` function first added?
  ```bash
  $ git log -S "getUserById"
  ```
  This will show you the commit that introduced that string.

- **By Range**:
  `git log main..feature`
  Shows all commits that are in `feature` but not in `main`.

### `git blame`: Who Wrote This Line?

`git blame <file>` is a code archaeology tool. It annotates every line in a file, showing which commit last modified it, who the author was, and when.

```bash
$ git blame README.md
^a1b2c3d (Alice 2023-03-15 10:00:00 -0700 1) This is the first line.
d4e5f6g (Bob   2023-03-16 14:30:00 -0700 2) This is the second line, added by Bob.
^a1b2c3d (Alice 2023-03-15 10:00:00 -0700 3) This is the third line.
```
- The `^` before a SHA indicates that this line was introduced in the file's initial commit.

### The Problem with `git blame`

`blame` can be misleading. If a developer runs a code formatter or does a large-scale refactoring, they will appear as the "author" of many lines they didn't originally write. `blame` only shows the last person to *touch* a line.

### Advanced `blame`: Ignoring Revisions

If you know a specific commit was just a big, noisy refactoring, you can tell `blame` to ignore it.

1.  Create a file, e.g., `.git-blame-ignore-revs`.
2.  Add the SHAs of the formatting/refactoring commits to this file, one per line.
3.  Run `blame` with the `--ignore-revs-file` flag.

```bash
$ git blame --ignore-revs-file .git-blame-ignore-revs README.md
```
Now, `blame` will skip over those commits and show you the *real* author of the line from before the refactoring. You can configure this in your `.gitconfig` to be used automatically.

### `blame` Archaeology

- `git blame -L 10,20 <file>`: Only show the blame for lines 10 through 20.
- `git blame -e`: Show the author's email address.
- `git blame -w`: Ignore whitespace changes.

**The "Walk Backwards" Workflow:**
You see a line and `blame` tells you it was changed in commit `d4e5f6g`. But you want to know what the line was *before* that.

1.  Run `git blame <file>`. Find the line and the commit `d4e5f6g`.
2.  Run `git show d4e5f6g` to see the change that was made.
3.  To see the state of the file just before that commit, run `git blame d4e5f6g^ -- <file>`. The `^` means "the parent of this commit". This will run `blame` on the file as it existed in the previous commit, allowing you to trace a line's history backwards through time.

### Key Takeaways

- Master `git log` formatting and filtering to quickly find the information you need. Create aliases for your favorite formats.
- Use `git log -S` (the pickaxe) to find when a specific piece of code was introduced or removed.
- `git blame` tells you which commit last modified each line of a file.
- Be aware that `blame` can be misleading after large refactorings. Use `--ignore-revs-file` to get a more accurate history.
- Use the "walk backwards" workflow with `blame` to trace the evolution of a specific line of code.

### Interview Notes

- **Question**: "I've found a bug in the code. How would you go about finding out when this bug was introduced?"
- **Answer**: "My first step would be to use `git blame` on the file with the bug to see who last touched the problematic lines and when. This gives me a starting commit to investigate. I'd then use `git show <commit-sha>` to see the context of that change. If that commit introduced the bug, I'm done. If the bug existed before that commit, I'd use the `blame` 'walk backwards' technique: `git blame <commit-sha>^ -- <file>` to see the state of the file in the parent commit. I'd repeat this process, walking back through history. If the history is long, a more powerful and automated approach would be to use `git bisect`, which can pinpoint the exact commit that introduced the bug much more quickly."
