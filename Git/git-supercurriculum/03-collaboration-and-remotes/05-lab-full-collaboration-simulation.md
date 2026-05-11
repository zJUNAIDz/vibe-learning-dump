
# 05-lab-full-collaboration-simulation.md

- **Purpose**: A hands-on lab simulating a multi-developer workflow, including feature development, a rejected push, resolving divergence, and handling a PR.
- **Estimated Difficulty**: 5/5
- **Estimated Time**: 90 minutes
- **Prerequisites**: All previous lessons in Module 03.

---

### Goal

To simulate a realistic collaborative workflow between two developers, "Alice" and "Bob," demonstrating how to manage shared branches and integrate work safely.

### Setup: The "Remote" Repository

First, we need a "central" repository to act as our `origin`. This will be a bare repository, which is a repository with no working directory, just the `.git` data.

```bash
# In your terminal, outside of any existing repo
$ mkdir collaboration-lab && cd collaboration-lab

# Create the central "remote" repo
$ git init --bare remote.git
```

### Setup: Developer Repositories

Now, let's create two local repositories, one for Alice and one for Bob, by cloning from our bare remote.

```bash
# Alice's repo
$ git clone remote.git alice
Cloning into 'alice'...
warning: You appear to have cloned an empty repository.
done.

# Bob's repo
$ git clone remote.git bob
Cloning into 'bob'...
warning: You appear to have cloned an empty repository.
done.
```
We now have a `remote.git` directory, an `alice` directory, and a `bob` directory.

### Part 1: Alice Starts a Feature

1.  **Alice creates the initial commit.**
    ```bash
    $ cd alice
    $ echo "Project v1" > README.md
    $ git add .
    $ git commit -m "Initial commit"
    $ git push origin main
    $ cd ..
    ```

2.  **Bob pulls the initial commit.**
    ```bash
    $ cd bob
    $ git pull origin main
    $ ls # Should see README.md
    ```

3.  **Alice starts a new feature.**
    ```bash
    $ cd ../alice
    $ git switch -c feature/add-login
    $ echo "Login Page" > login.html
    $ git add .
    $ git commit -m "Feat: Add login page structure"
    $ git push -u origin feature/add-login
    $ cd ..
    ```

### Part 2: Work Diverges

Now, both Alice and Bob will work on the `feature/add-login` branch simultaneously.

1.  **Bob checks out the feature branch and starts working.**
    ```bash
    $ cd bob
    $ git switch feature/add-login
    $ echo "function login() {}" > js/login.js
    $ git add .
    $ git commit -m "Feat: Add login script skeleton"
    $ cd ..
    ```

2.  **At the same time, Alice also works on the branch.**
    ```bash
    $ cd alice
    $ echo "<!-- Add form here -->" >> login.html
    $ git add .
    $ git commit -m "Feat: Add placeholder for login form"
    $ cd ..
    ```

3.  **Alice pushes first.** Her push will succeed because it's a fast-forward.
    ```bash
    $ cd alice
    $ git push
    $ cd ..
    ```

4.  **Bob tries to push.** His push will be **rejected**.
    ```bash
    $ cd bob
    $ git push
    # Observe the "non-fast-forward" rejection error.
    ```

### Part 3: Bob Resolves the Divergence

Bob must now integrate the changes from the remote before he can push. He will use the `pull --rebase` strategy.

1.  **Bob pulls with rebase.**
    ```bash
    $ cd bob
    $ git pull --rebase origin feature/add-login
    # The output should show it's rebasing his commit on top of Alice's.
    # First, rewinding head to replay your work on top of it...
    # Applying: Feat: Add login script skeleton
    ```

2.  **Bob inspects the history.**
    ```bash
    $ git log --oneline --graph
    # The history should be linear. Alice's commit should be first,
    # followed by Bob's rebased commit.
    # * 1a2b3c4 (HEAD -> feature/add-login) Feat: Add login script skeleton
    # * 4d5e6f7 (origin/feature/add-login) Feat: Add placeholder for login form
    # * ...
    ```

3.  **Bob pushes his work.** This time, it will be a fast-forward and will succeed.
    ```bash
    $ git push
    $ cd ..
    ```

### Part 4: Alice Opens a Pull Request

Alice is now ready to merge the feature.

1.  **Alice updates her local branch.** She needs Bob's latest commit.
    ```bash
    $ cd alice
    $ git switch feature/add-login
    $ git pull # A simple pull/merge is fine here as it will be a fast-forward
    ```

2.  **Alice cleans up the history (Optional but good practice).** The history has two commits. Let's say the team policy is to have one commit per feature. Alice will do an interactive rebase to squash them.
    ```bash
    $ git rebase -i origin/main
    # In the interactive editor, change the second commit from 'pick' to 'squash'.
    # Save and exit.
    # Edit the new combined commit message to be "Feat: Add login page".
    ```

3.  **Alice force-pushes her cleaned-up branch.** Because she rewrote history, she must use `--force-with-lease`.
    ```bash
    $ git push --force-with-lease
    ```

4.  **Simulate the PR merge.** In a real scenario, Alice would open a PR on GitHub/GitLab. Here, we'll simulate the merge directly.
    ```bash
    $ git switch main
    $ git pull origin main # Make sure main is up-to-date
    $ git merge --no-ff feature/add-login
    $ git push origin main
    ```

### Debrief

This lab simulated a complete, realistic workflow:
- Setting up a shared remote.
- Simultaneous work on a shared feature branch.
- A rejected push due to divergent history.
- Resolving the divergence using `pull --rebase` for a clean history.
- Cleaning up a feature branch's commits using interactive rebase before merging.
- Safely force-pushing a rebased branch.
- Merging the completed feature into `main`.

This cycle of `code -> pull --rebase -> push` is the fundamental rhythm of collaborative Git development on shared branches.
