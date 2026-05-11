
# 05-lab-branching-strategy-simulation.md

- **Purpose**: A hands-on lab to simulate managing a project using the GitFlow methodology, including features, a release, and a hotfix.
- **Estimated Difficulty**: 5/5
- **Estimated Time**: 90 minutes
- **Prerequisites**: All previous lessons in Module 06.

---

### Goal

To gain practical experience with the GitFlow branching model by simulating a full development and release cycle.

### Setup

We'll start by creating a repository with the two primary GitFlow branches: `main` and `develop`.

```bash
$ mkdir gitflow-lab && cd gitflow-lab
$ git init

# Create the main branch and initial commit
$ echo "Version 1.0" > version.txt && git add . && git commit -m "Initial commit"
$ git tag v1.0.0

# Create the develop branch based on main
$ git switch -c develop
```
Your repository is now set up. `main` represents the stable v1.0.0 release, and `develop` is ready for new work.

### Part 1: Developing a New Feature

You need to add a new feature.

1.  **Create a feature branch from `develop`.**
    ```bash
    $ git switch develop
    $ git switch -c feature/add-changelog
    ```
2.  **Work on the feature.**
    ```bash
    $ echo "Changelog:" > CHANGELOG.md
    $ git add . && git commit -m "Feat: Add changelog file"
    ```
3.  **Finish the feature and merge it back to `develop`.** (Simulating a PR merge).
    ```bash
    $ git switch develop
    $ git merge --no-ff feature/add-changelog
    $ git branch -d feature/add-changelog
    ```
`develop` now contains the new feature and is ahead of `main`.

### Part 2: Preparing a Release

The `develop` branch is now considered "feature complete" for version 1.1.0. It's time to create a release branch to prepare for deployment.

1.  **Create a `release` branch from `develop`.**
    ```bash
    $ git switch develop
    $ git switch -c release/v1.1.0
    ```
2.  **Perform release-specific tasks.** While on the release branch, you bump the version number.
    ```bash
    $ echo "Version 1.1.0" > version.txt
    $ git add . && git commit -m "Bump version to 1.1.0"
    ```
    Imagine that during final testing on this branch, a small bug is found and fixed here as well.
    ```bash
    $ echo "Changelog v1.1" > CHANGELOG.md
    $ git add . && git commit -m "Fix: Correct changelog content for release"
    ```

### Part 3: Finalizing the Release

The `release/v1.1.0` branch is now stable and ready.

1.  **Merge the release into `main`.**
    ```bash
    $ git switch main
    $ git merge --no-ff release/v1.1.0
    ```
2.  **Tag the new release on `main`.**
    ```bash
    $ git tag -a v1.1.0 -m "Release version 1.1.0"
    ```
3.  **Merge the release back into `develop`.** This is critical! The version bump and bug fix need to be in `develop`.
    ```bash
    $ git switch develop
    $ git merge --no-ff release/v1.1.0
    ```
4.  **Delete the release branch.**
    ```bash
    $ git branch -d release/v1.1.0
    ```
The release is complete. `main` and `develop` both contain the v1.1.0 code.

### Part 4: The Emergency Hotfix

A critical bug is discovered in the v1.1.0 release that is now in production.

1.  **Create a `hotfix` branch from `main`.** It must be based on the production code.
    ```bash
    $ git switch main
    $ git switch -c hotfix/critical-bug
    ```
2.  **Fix the bug.**
    ```bash
    $ echo "Version 1.1.1 - Hotfix" > version.txt
    $ git add . && git commit -m "Fix: Critical production issue"
    ```
3.  **Finalize the hotfix.**
    - **Merge to `main` and tag:**
      ```bash
      $ git switch main
      $ git merge --no-ff hotfix/critical-bug
      $ git tag -a v1.1.1 -m "Hotfix for critical issue"
      ```
    - **Merge back to `develop`:**
      ```bash
      $ git switch develop
      $ git merge --no-ff hotfix/critical-bug
      ```
4.  **Delete the hotfix branch.**
    ```bash
    $ git branch -d hotfix/critical-bug
    ```

### Debrief

Examine your commit history graph.

```bash
$ git log --graph --oneline --decorate --all
```
You should see a clear, structured history that tells the story of your project:
- The `main` branch has a linear history of tagged releases (`v1.0.0`, `v1.1.0`, `v1.1.1`).
- The `develop` branch contains the full history, including feature work and the merges from the release and hotfix branches.
- The temporary branches (`feature`, `release`, `hotfix`) are gone, but their work is preserved correctly in the two main branches.

This lab demonstrates the discipline and structure of GitFlow. While it involves many steps, each step has a clear purpose, resulting in a predictable and stable release process.
