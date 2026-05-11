
# 01-git-aliases-for-speed-and-efficiency.md

- **Purpose**: To demonstrate how to create Git aliases to shorten common commands, improve workflow, and build complex new commands.
- **Estimated Difficulty**: 2/5
- **Estimated Reading Time**: 30 minutes
- **Prerequisites**: Basic command-line usage.

---

### What are Git Aliases?

An alias is a custom shortcut you can define for a Git command. If you find yourself typing the same long command over and over, you can create a short, memorable alias for it.

This is one of the simplest and most powerful ways to improve your productivity with Git. Aliases are stored in your Git configuration file (`~/.gitconfig` for global aliases, or `.git/config` for repository-specific ones).

### How to Create an Alias

You create aliases using the `git config` command with the `alias.<alias-name>` key.

```bash
$ git config --global alias.co checkout
$ git config --global alias.br branch
$ git config --global alias.ci commit
$ git config --global alias.st status
```

Now, instead of typing `git checkout`, you can just type `git co`. Instead of `git status`, you can type `git st`.

This might seem like a small saving, but over the course of a day, these keystrokes add up. It also reduces the cognitive load of remembering long command names.

### Practical, Powerful Aliases

Simple shortcuts are just the beginning. The real power of aliases comes from creating new, more complex commands.

**1. The "Prettylog" Alias**
`git log` is powerful, but its default output is verbose. A common practice is to create a custom log format that is dense and readable.

```bash
$ git config --global alias.lg "log --graph --pretty=format:'%C(yellow)%h%C(reset) -%C(red)%d%C(reset) %s %C(green)(%cr) %C(bold blue)<%an>%C(reset)' --abbrev-commit"
```
Now, `git lg` gives you a beautiful, one-line graph view of your history.

**2. The "Last Commit" Alias**
How often do you type `git log -1 HEAD` to see the last commit?

```bash
$ git config --global alias.last "log -1 HEAD"
```
Now, just `git last`.

**3. The "Unstage" Alias**
The command to unstage a file is `git restore --staged <file>`. This is long and not very intuitive.

```bash
$ git config --global alias.unstage "restore --staged"
```
Now you can run `git unstage my-file.txt`, which is much clearer.

### Shell Aliases: Taking it to the Next Level

Sometimes, you want to create an alias that isn't just a Git command, but a full shell command. You can do this by starting the alias with an `!`. Git will pass the rest of the command to the shell.

**1. The "Assume Unchanged" Alias**
Sometimes you have a tracked file that you want to temporarily modify locally but never commit (e.g., a config file pointing to a local database).

```bash
$ git config --global alias.assume-unchanged "update-index --assume-unchanged"
$ git config --global alias.unassume-unchanged "update-index --no-assume-unchanged"
```
Now you can `git assume-unchanged config.local.js`.

**2. The "List Aliases" Alias**
How do you see all the aliases you've created?

```bash
# This alias runs a shell command to grep your config for all alias definitions.
$ git config --global alias.aliases "! git config --get-regexp ^alias\. | sed -e s/^alias\.// -e s/\ /\ =\ /"
```
Now, `git aliases` will print a neat list of all your custom commands.

**3. The "Delete Merged Branches" Alias**
A common cleanup task is to delete local branches that have already been merged into `main`.

```bash
$ git config --global alias.cleanup "! git branch --merged main | grep -v '^[ *]*main$' | xargs git branch -d"
```
This complex command:
- `git branch --merged main`: Lists all branches merged into `main`.
- `grep -v '^[ *]*main$'`: Filters out the `main` branch itself from the list.
- `xargs git branch -d`: Takes the resulting list of branch names and runs `git branch -d` on each one.

Now, a simple `git cleanup` will tidy up your local repository.

### Sharing Aliases

Aliases are part of your Git configuration. They are not stored in the repository, so they are not shared with your team by default.

- **For personal productivity**: Keep them in your global `~/.gitconfig` file.
- **For team-wide aliases**: You can create a script in your repository that team members can run to install a set of shared aliases into their local config. However, it's often better to use actual shell scripts checked into the repository for complex, shared commands.

### My Favorite Aliases: A Recommended Set

Here is a good starting set for any developer.

```bash
# Simple shortcuts
git config --global alias.co checkout
git config --global alias.br branch
git config --global alias.ci commit
git config --global alias.st status

# Staging
git config --global alias.a "add -A"

# Resetting
git config --global alias.unstage "reset HEAD --"
git config --global alias.undo "reset --hard HEAD~1"

# History
git config --global alias.lg "log --graph --oneline --decorate"
git config --global alias.last "log -1 HEAD"

# Branches
git config --global alias.cleanup "! git branch --merged main | grep -v '^[ *]*main$' | xargs git branch -d"
```

### Key Takeaways

- Git aliases are custom shortcuts for Git commands.
- They are a simple but powerful way to improve your speed and reduce cognitive load.
- Use `git config --global alias.<name> "<command>"` to create them.
- Use `!` to create shell aliases that can run complex scripts and pipe commands together.
- Create aliases for your most frequent operations and for complex commands you don't want to memorize.
