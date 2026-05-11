
# 04-release-and-hotfix-workflows.md

- **Purpose**: To provide a practical guide to managing releases and hotfixes, showing how these workflows are implemented within different branching strategies.
- **Estimated Difficulty**: 4/5
- **Estimated Reading Time**: 40 minutes
- **Prerequisites**: `01-gitflow-a-structured-approach.md`, `02-github-flow-a-lightweight-alternative.md`

---

### What is a Release?

A "release" is a version of your software that is delivered to end-users. The process of creating and managing these releases is a critical part of the software development lifecycle. Git is the central tool for managing the code that goes into a release.

### What is a Hotfix?

A "hotfix" is an emergency fix for a critical bug in a production release. The hotfix workflow needs to be fast, isolated, and reliable to resolve the production issue without disrupting ongoing feature development.

### Workflow 1: Release Management in GitFlow

GitFlow has a highly structured, built-in process for releases and hotfixes.

**The Release Workflow:**

1.  **Create a `release` branch**: When `develop` is ready for the next release, a branch is created from it.
    `git switch develop`
    `git pull`
    `git switch -c release/v2.0.0`

2.  **Stabilize the release**: This branch is now in a feature freeze. Only bug fixes, documentation generation, and other release-oriented tasks should happen here. Commits are made directly to this branch.

3.  **Finalize the release**: When the release branch is stable and ready, it is merged into `main` and `develop`.
    - **Merge to `main` and tag**:
      `git switch main`
      `git merge --no-ff release/v2.0.0`
      `git tag -a v2.0.0 -m "Release version 2.0.0"`
      `git push origin main --tags`

    - **Merge back to `develop`**: This is a crucial step to ensure any fixes made on the release branch get back into the main line of development.
      `git switch develop`
      `git merge --no-ff release/v2.0.0`
      `git push origin develop`

4.  **Delete the release branch**:
    `git branch -d release/v2.0.0`

**The Hotfix Workflow:**

1.  **Create a `hotfix` branch from `main`**: The branch must be based on the production code.
    `git switch main`
    `git pull`
    `git switch -c hotfix/fix-critical-bug main`

2.  **Fix the bug**: Make the necessary code changes and commit them.
    `git commit -m "Fix: Resolve critical bug X"`

3.  **Finalize the hotfix**: Similar to a release, the hotfix must be merged into both `main` and `develop`.
    - **Merge to `main` and tag**:
      `git switch main`
      `git merge --no-ff hotfix/fix-critical-bug`
      `git tag -a v2.0.1 -m "Hotfix for critical bug X"`
      `git push origin main --tags`

    - **Merge back to `develop`**:
      `git switch develop`
      `git merge --no-ff hotfix/fix-critical-bug`
      `git push origin develop`

4.  **Delete the hotfix branch**.

### Workflow 2: Release Management in GitHub Flow / TBD

These models don't have a concept of a "release" in the same way. The `main` branch *is* the release. However, you still need to manage versions and support older releases.

**The Release Workflow (Creating a "Support" Branch):**

In these models, you don't branch to *create* a release; you branch from a release to *support* it.

1.  **Tag a commit on `main`**: When you decide that the current state of `main` constitutes a new major release, you tag it.
    `git switch main`
    `git pull`
    `git tag -a v2.0.0 -m "Release version 2.0.0"`
    `git push origin main --tags`

2.  **Continue development on `main`**: `main` continues to move forward with new features and fixes.

3.  **Create a support branch (only if needed)**: A bug is found in `v2.0.0`, but `main` has already changed significantly. You need to patch `v2.0.0` without including all the new features. You now create a branch from the tag.
    `git switch -c support/v2.0 v2.0.0`
    This `support/v2.0` branch is now a long-lived branch for patching the `2.0` release line.

**The Hotfix Workflow:**

1.  **Fix the bug on the support branch**:
    `git switch support/v2.0`
    `git commit -m "Fix: Resolve bug Y"`

2.  **Tag the new patch release**:
    `git tag -a v2.0.1 -m "Patch release for bug Y"`
    `git push origin support/v2.0 --tags`

3.  **Backport the fix to `main`**: The bug likely also exists in `main`. You need to bring the fix forward. This is a perfect use case for `cherry-pick`.
    `git switch main`
    `git cherry-pick <sha_of_fix_commit_on_support_branch>`
    `git push origin main`

### Comparison of Approaches

| Aspect          | GitFlow                                               | GitHub Flow / TBD                                     |
| --------------- | ----------------------------------------------------- | ----------------------------------------------------- |
| **Release Prep**| Done on a dedicated `release` branch before merging to `main`. | `main` is always ready. A `support` branch is created from `main` *after* release if needed. |
| **Hotfix Origin** | Branched from `main`.                                 | Branched from a `support` branch (or `main` if no support branch exists). |
| **Hotfix Integration** | Merged into both `main` and `develop`.            | Fixed on the `support` branch, then cherry-picked to `main`. |
| **Complexity**  | Higher. More branches and merges to manage.           | Lower. Fewer long-lived branches.                     |
| **Best For**    | Projects with a clear, scheduled release cadence.     | Projects with continuous deployment and a need to support multiple past versions. |

### Key Takeaways

- Release and hotfix workflows are formalized processes for managing the state of production code.
- **GitFlow** is prescriptive, with built-in `release` and `hotfix` branches that get merged into both `main` and `develop`.
- **GitHub Flow / TBD** are more flexible. Releases are tags on `main`. Hotfixes are done on long-lived `support` branches created from those tags and then cherry-picked back into `main`.
- The choice of workflow depends entirely on your project's release strategy.
