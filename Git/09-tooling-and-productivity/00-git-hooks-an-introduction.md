
# 00-git-hooks-an-introduction.md

- **Purpose**: To introduce Git Hooks as a mechanism for triggering custom scripts in response to Git events, enabling automation and policy enforcement.
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 30 minutes
- **Prerequisites**: Basic shell scripting knowledge.

---

### What are Git Hooks?

Git Hooks are scripts that Git executes automatically before or after certain events, such as `commit`, `push`, and `receive`. They are a built-in feature of Git that lets you customize its behavior and enforce rules.

Hooks are stored in the `.git/hooks` directory of every Git repository. When you initialize a repository, Git populates this directory with a number of example scripts, all ending in `.sample`. To enable a hook, you just need to remove the `.sample` extension and make the script executable.

### Client-Side vs. Server-Side Hooks

There are two main categories of hooks:

1.  **Client-Side Hooks**:
    - These are triggered by operations on a developer's local repository.
    - They are used to enforce policies on the developer's own workflow.
    - Examples: `pre-commit`, `prepare-commit-msg`, `commit-msg`, `post-commit`, `pre-push`.
    - **Important**: Client-side hooks are **not** shared when you clone a repository. Each developer must set them up on their own machine. This makes them difficult to enforce across a team.

2.  **Server-Side Hooks**:
    - These run on the remote repository (e.g., on your GitHub Enterprise or self-hosted GitLab server).
    - They are used to enforce policies for the entire project.
    - Examples: `pre-receive`, `update`, `post-receive`.
    - These are the authoritative source for enforcing rules, as they cannot be bypassed by developers.

### Common Client-Side Hooks

**1. `pre-commit`**
- **Triggered**: Before a commit is created.
- **Use Case**: This is the most common hook. It's used to run checks on the code that is about to be committed.
    - Run a linter (`eslint`, `flake8`).
    - Run a code formatter (`prettier`, `black`).
    - Run fast unit tests.
    - Check for debug statements like `console.log`.
- **Mechanism**: If the `pre-commit` script exits with a non-zero status, the `git commit` command is aborted.

**Example `pre-commit` script:**
```bash
#!/bin/sh
echo "Running linter..."
npm run lint
if [ $? -ne 0 ]; then
  echo "Linting failed. Aborting commit."
  exit 1
fi
exit 0
```

**2. `commit-msg`**
- **Triggered**: After you've written a commit message, but before the commit is created.
- **Use Case**: To enforce a specific commit message format (e.g., Conventional Commits).
- **Mechanism**: The script receives the path to the file containing the commit message. It can read and validate this file. If the script exits non-zero, the commit is aborted.

**Example `commit-msg` script:**
```bash
#!/bin/sh
MSG_FILE=$1
MSG=$(cat $MSG_FILE)

# Check if the message starts with 'feat:' or 'fix:'
if ! echo "$MSG" | grep -qE "^(feat|fix|docs|style|refactor|test|chore)(\(.+\))?:"; then
  echo "ERROR: Commit message does not follow Conventional Commits format."
  exit 1
fi
exit 0
```

**3. `pre-push`**
- **Triggered**: Before a `git push` operation.
- **Use Case**: To run final, more comprehensive checks before sharing your code with others.
    - Run the full test suite (not just the fast unit tests).
    - Ensure the branch is up-to-date with the remote.
- **Mechanism**: If the script exits non-zero, the `push` is aborted.

### The Problem with Client-Side Hooks

The biggest issue with client-side hooks is that they are not cloned with the repository. You can't guarantee that every developer on your team has them installed or hasn't modified them. A developer can always bypass a client-side hook by running `git commit --no-verify`.

### Frameworks for Managing Hooks

Because managing hooks manually is difficult, several popular frameworks have emerged to simplify the process.

- **Husky**: A very popular tool in the JavaScript ecosystem. It allows you to define your Git hooks in your `package.json` file. When a developer runs `npm install`, Husky automatically sets up the hooks in their `.git/hooks` directory.
- **pre-commit**: A language-agnostic framework. It manages a `.pre-commit-config.yaml` file where you define a list of hooks you want to run. It's powerful and supports a wide range of common checks.

These tools solve the sharing problem by making the hook configuration part of the repository itself.

### Server-Side Hooks: The Enforcers

Server-side hooks are the ultimate authority. Since they run on the central server, they cannot be bypassed.

**`pre-receive`**
- **Triggered**: When a client pushes to the remote repository, `pre-receive` runs *before* any refs are updated.
- **Use Case**: This is the most common server-side hook for enforcing policy.
    - Reject commits that don't follow the commit message format.
    - Reject pushes that are not fast-forwards (i.e., block force pushes on `main`).
    - Check if the author's email matches a list of approved contributors.
    - Trigger a pre-emptive CI build to see if the changes will pass tests *before* accepting the push.
- **Mechanism**: The script receives a list of all the refs being pushed. If it exits non-zero, the entire push is rejected.

Most developers will not write server-side hooks themselves. Instead, they will configure them through the web UI of their Git hosting platform (e.g., GitHub's "branch protection rules"). These rules are essentially a user-friendly interface for managing `pre-receive` hooks.

### Key Takeaways

- Git Hooks are scripts that run automatically on Git events.
- **Client-side hooks** (`pre-commit`, `commit-msg`) are for individual developer convenience and workflow automation. They are not enforceable.
- **Server-side hooks** (`pre-receive`) are for enforcing project-wide policies. They are the source of truth.
- Use frameworks like **Husky** or **pre-commit** to easily share and manage client-side hooks across a team.
- Use your hosting platform's **branch protection rules** to configure server-side policies like requiring CI checks to pass before merging.
