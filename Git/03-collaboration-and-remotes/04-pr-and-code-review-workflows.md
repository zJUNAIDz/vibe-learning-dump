
# 04-pr-and-code-review-workflows.md

- **Purpose**: To discuss common code review workflows, focusing on how Git is used to manage feedback and revisions in a Pull Request (or Merge Request).
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 40 minutes
- **Prerequisites**: `01-the-push-and-pull-dance.md`

---

### The Pull Request (PR)

A Pull Request (PR) is not a Git feature. It's a feature of hosting platforms like GitHub, GitLab, and Bitbucket built *on top* of Git. It's a formal proposal to merge changes from one branch into another.

The core of a PR is a request to merge a `head` branch (your feature branch) into a `base` branch (e.g., `main`). The platform provides a web UI for:
- Discussing the changes.
- Commenting on specific lines of code.
- Integrating with CI/CD checks (e.g., tests, linting).
- Ultimately, performing the merge.

### The Code Review Cycle

The typical cycle looks like this:
1.  You push your feature branch and open a PR.
2.  Your teammates review the code and leave comments with requested changes.
3.  You make the requested changes locally.
4.  You add the new changes to your feature branch.
5.  The PR is updated automatically, and the cycle repeats until the PR is approved.

The key question is step 4: **How do you add the new changes to your feature branch?** There are two main strategies.

### Strategy 1: The "Amendment" Workflow (Rebase)

This strategy focuses on keeping the commit history of the feature branch clean and atomic. Each commit should be a perfect, self-contained change. When you get feedback, you don't add new commits like "Fix typo" or "Address PR comments." Instead, you amend the existing commits.

**Workflow:**
1.  You have a PR with one or more commits. A reviewer asks for a change in a file that was part of your first commit.
2.  Locally, you make the change.
3.  You stage the change: `git add <file>`.
4.  You use interactive rebase (`git rebase -i main`) to edit your history.
5.  You find the commit you want to change and mark it with `edit`.
6.  When the rebase pauses at that commit, you add your change to it: `git commit --amend --no-edit`.
7.  You continue the rebase: `git rebase --continue`.
8.  Since you have rewritten history, you must force-push to your feature branch: `git push --force-with-lease origin <feature-branch>`.

**Pros:**
- The final history in `main` is pristine. Each commit is a single, logical unit.
- It's easy to see the final state of the feature.

**Cons:**
- **Loses review context.** Force-pushing rewrites the history of the PR. On GitHub/GitLab, this can make it hard to see what changed between review rounds. Some developers find it disorienting to have the history they were just commenting on disappear.
- Requires a strong understanding of interactive rebase.
- Can be complex if feedback applies to multiple commits.

### Strategy 2: The "Append-Only" Workflow (Merge)

This strategy treats the PR branch as a chronological record of the review process. All changes, including fixes from feedback, are added as new commits.

**Workflow:**
1.  You have a PR. A reviewer asks for changes.
2.  Locally, you make the changes.
3.  You create a new commit: `git commit -m "Fix: Address review comments from Alice"`.
4.  You push the new commit to your feature branch: `git push origin <feature-branch>`.

The PR is updated with the new commit. The reviewer can now easily see exactly what you changed in response to their feedback.

Once the PR is approved, you have a choice at the time of merging:
- **Merge Commit (`git merge --no-ff`)**: This preserves the entire history, including all the "fixup" commits. The history shows the full story of the feature's development and review.
- **Squash and Merge**: This is the most common approach with this workflow. The hosting platform will take all the commits in the PR (`Initial feature`, `Fix typo`, `Address comments`) and squash them into a single, perfect commit before merging it into `main`.

**Pros:**
- **Preserves review context.** It's very easy to see what changed between review rounds.
- Simpler for developers; no rebasing or force-pushing required.

**Cons:**
- The feature branch history can become very noisy with lots of small commits.
- Relies on "Squash and Merge" to keep the `main` branch history clean. If you don't squash, `main` gets cluttered.

### Which is Better?

There is no single "best" way. It's a team-level decision.

- The **Amendment/Rebase** workflow is often preferred by those who prioritize a perfect, logical commit history above all else. It's common in open-source projects (like Git itself).
- The **Append-Only/Squash** workflow is often preferred by teams who prioritize the clarity and context of the code review process itself. It's arguably more common in corporate/product engineering teams.

The most important thing is that the **entire team agrees on one strategy.** Mixing them causes confusion.

### Local PR Checkout

Many platforms have CLI tools to make reviewing easier. For example, with the GitHub CLI (`gh`):

```bash
# Check out PR #123 locally
$ gh pr checkout 123
```
This will fetch the branch from the contributor's fork and switch to it, making it trivial to test a PR locally. This is often a great use case for `git worktree`.

### Key Takeaways

- A PR is a platform feature for discussing and merging Git branches.
- The two main strategies for updating a PR are:
    1.  **Rebase/Amend**: Rewriting history to keep commits clean, requires `force-push`.
    2.  **Append/Squash**: Adding new commits to preserve review context, then squashing at the end.
- The choice of strategy is a team convention. Consistency is key.
- Use platform CLI tools or `worktree` to easily check out PRs for local review.
