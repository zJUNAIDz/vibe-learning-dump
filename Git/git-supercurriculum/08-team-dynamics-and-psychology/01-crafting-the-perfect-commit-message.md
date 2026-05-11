
# 01-crafting-the-perfect-commit-message.md

- **Purpose**: To establish a clear and effective standard for writing Git commit messages that are useful, readable, and professional.
- **Estimated Difficulty**: 2/5
- **Estimated Reading Time**: 30 minutes
- **Prerequisites**: Basic `git commit` usage.

---

### Why Do Commit Messages Matter?

A commit message is a piece of documentation. It tells a story. When you run `git log`, `git blame`, or review a pull request, the commit messages are the primary way you understand the *why* behind a change.

- **Good messages** help your teammates (and your future self) understand the context of your code without having to read the code itself. They are essential for debugging and archaeology.
- **Bad messages** (`"wip"`, `"fix bug"`, `"stuff"`) are useless. They are a missed opportunity to communicate and document.

Investing a few extra minutes in writing a good commit message is one of the highest-leverage activities a developer can do.

### The Anatomy of a Great Commit

The most widely adopted standard for commit messages follows this structure:

1.  **Subject Line**: A short, imperative summary of the change.
2.  **Blank Line**: A single blank line separating the subject from the body.
3.  **Body (Optional)**: A more detailed explanation of the change.
4.  **Footer (Optional)**: References to issue trackers or pull requests.

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Rule 1: The Subject Line

The subject line is the most important part. It's what you see in `git log --oneline`.

- **Use the imperative mood**: "Add feature" not "Added feature" or "Adds feature". Think of it as giving a command to the codebase: `(If applied, this commit will) Add feature...`. This is also the style Git itself uses for messages (e.g., `Merge branch '...'`).
- **Capitalize the first letter.**
- **Do not end with a period.**
- **Limit to 50 characters.** This is a soft limit, but it forces you to be concise. It also ensures messages look good in a variety of Git tools.
- **Use a Prefix (Conventional Commits)**: The "Conventional Commits" specification is a popular standard that adds a `type` and optional `scope` to the subject line.
    - `feat:`: A new feature.
    - `fix:`: A bug fix.
    - `docs:`: Documentation only changes.
    - `style:`: Code style changes (formatting, etc.).
    - `refactor:`: A code change that neither fixes a bug nor adds a feature.
    - `test:`: Adding missing tests or correcting existing tests.
    - `chore:`: Changes to the build process or auxiliary tools.

**Examples of Good Subject Lines:**
- `feat(auth): Add password reset functionality`
- `fix(api): Correct pagination error on user endpoint`
- `docs(readme): Update installation instructions`
- `refactor: Simplify user service caching logic`

### Rule 2: The Blank Line

You must include a single blank line between the subject and the body. Many Git tools (including `git log` and web UIs) use this blank line to distinguish the subject from the body. If you omit it, the entire message may be treated as the subject line.

### Rule 3: The Body

The body is where you explain the **why**, not the **how**. The code itself shows *how* the change was made. The body should explain the context.

- **Explain the problem**: What was the state of the world before this change? What bug was occurring? What feature was missing?
- **Explain your solution**: Why did you choose this particular approach? What alternatives did you consider?
- **Wrap the body at 72 characters.** This makes the log much more readable in a standard terminal.

**Example Body:**
```
The user endpoint was previously returning all users without
considering the 'limit' query parameter. This could cause
performance issues and excessive data transfer for clients.

This change implements server-side pagination by respecting
the 'limit' and 'offset' parameters, falling back to a
default limit of 100 if not provided.
```

### Rule 4: The Footer

The footer is used for machine-readable metadata. The most common use is for linking to issue trackers.

- `Fixes: #123`
- `Closes: #456, #789`
- `Refs: #246`

Many platforms like GitHub will automatically parse these footers. For example, `Fixes: #123` will automatically close issue #123 when the commit is merged into the `main` branch.

### A Complete Example

```
fix(api): Correct pagination error on user endpoint

The user endpoint was previously returning all users without
considering the 'limit' query parameter. This could cause
performance issues and excessive data transfer for clients.

This change implements server-side pagination by respecting
the 'limit' and 'offset' parameters, falling back to a
default limit of 100 if not provided.

Fixes: #123
```

### How to Enforce This

- **Editor Configuration**: Configure your editor to help you follow these rules (e.g., with plugins that highlight line length).
- **Commit Hooks**: Use a pre-commit hook (e.g., with a tool like `commitlint`) to check the message format before the commit is even created. This is the most effective way to ensure consistency across a team.

### Key Takeaways

- A good commit message is a crucial piece of documentation.
- Follow the **Subject, Blank Line, Body, Footer** structure.
- Write the subject in the **imperative mood** and keep it short.
- Use the body to explain the **why**, not the how.
- Use prefixes like `feat:` and `fix:` (Conventional Commits) to make your history more structured and machine-readable.
- Use footers to link to issues.
