
# 02-managing-technical-debt-in-git.md

- **Purpose**: To discuss how to use Git as a tool for identifying, documenting, and managing technical debt.
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 30 minutes
- **Prerequisites**: `01-crafting-the-perfect-commit-message.md`

---

### What is Technical Debt?

Technical debt is the implied cost of rework caused by choosing an easy (limited) solution now instead of using a better approach that would take longer. It's a metaphor. Just like financial debt, technical debt accrues "interest" – the longer you leave it, the more expensive it becomes to fix.

Not all technical debt is bad. Sometimes, you intentionally take on debt to meet a deadline, with a plan to pay it back later. The danger is in unintentional or unmanaged debt.

### Using Git to Document Technical Debt

Your Git history is a ledger of all the decisions made in your project. It's an excellent place to make technical debt explicit.

**1. The `TODO` and `FIXME` Comments**
The simplest method is to leave comments directly in the code.

```javascript
// TODO: This is a temporary solution. We should refactor this
// to use the new UserService once it's available.
function getLegacyUser(id) {
  // ...
}

// FIXME: This algorithm is O(n^2) and will not scale.
// We need to replace it with a hash-based approach.
function findMatchingUsers(users) {
  // ...
}
```
- **`TODO`**: Describes something that is missing or incomplete.
- **`FIXME`**: Describes something that is broken or suboptimal.

**Pros**: The debt is documented right next to the code it affects.
**Cons**: These comments can get lost. There's no easy way to get a project-wide view of all the debt.

**2. The "Tech Debt" Commit Message**
When you knowingly introduce technical debt, be honest about it in the commit message.

```
feat(payment): Add temporary Stripe integration

This is a quick integration to meet the launch deadline. It
does not handle recurring payments or multiple currencies,
and the error handling is minimal.

A follow-up task has been created to build a more robust
payment abstraction layer.

Refs: TICKET-456
```
This creates a permanent, searchable record of the decision. Someone running `git blame` on this code in the future will immediately understand the context.

### Using Git to Track and Prioritize Debt

**1. Grepping for `TODO`s**
You can use `git grep` to find all instances of `TODO` or `FIXME` in your repository.

```bash
$ git grep "TODO"
$ git grep "FIXME"
```
This can be part of a regular team process (e.g., "Tech Debt Friday") where the team reviews the output of this command and decides what to tackle.

**2. Tagging Debt with `git notes`**
`git notes` is a lesser-known feature that allows you to attach extra information to a commit object without changing its SHA. This is a great way to tag a commit as containing technical debt *after the fact*.

**Scenario**: You discover that a commit from six months ago introduced a significant performance issue.

```bash
# Find the commit
$ git log --oneline
# ...
# a1b2c3d Some old feature

# Add a note to it
$ git notes add -m "TECH_DEBT: This introduced an N+1 query problem in the user loader." a1b2c3d
```
Now, when you view the log, you can ask Git to show the notes as well.

```bash
$ git log --show-notes a1b2c3d
# ...
#
# Notes:
# TECH_DEBT: This introduced an N+1 query problem...
```
You can then search for all commits with a specific note:
`git log --grep="TECH_DEBT"`

This creates a structured way to track debt without rewriting history.

**3. The "Debt-Payoff" Branch**
For larger refactoring efforts, create a specific branch.
`git switch -c refactor/remove-jquery`

This isolates the work and makes the intent clear. The pull request for this branch provides an excellent opportunity to discuss the value of paying down the debt. The size of the PR diff (`+100 -2500`) can be a powerful motivator, showing how much code was simplified or removed.

### A Culture of Managing Debt

Tools and processes are helpful, but the most important thing is creating a team culture that acknowledges and prioritizes managing technical debt.

- **Blameless Post-mortems**: When a bug is caused by old, crusty code, the focus should be on "How can we improve this code?" not "Who wrote this?"
- **Allocate Time**: Good teams explicitly allocate a percentage of their time (e.g., 15-20%) in each sprint or cycle to non-feature work, including bug fixes, refactoring, and paying down technical debt.
- **Make it Visible**: Use the Git techniques above, along with tickets in your issue tracker, to make the debt visible. If it's not visible, it will never be prioritized.

### Key Takeaways

- Technical debt is a natural part of software development; the goal is to manage it, not eliminate it entirely.
- Use your Git history as a tool for documenting debt.
- Be explicit about knowingly incurring debt in your commit messages.
- Use `git grep` to find `TODO` and `FIXME` comments.
- Consider using `git notes` to tag commits with debt information after the fact.
- Foster a team culture that values paying down debt and allocates time for it.
