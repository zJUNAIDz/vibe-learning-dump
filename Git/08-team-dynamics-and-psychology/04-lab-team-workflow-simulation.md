
# 04-lab-team-workflow-simulation.md

- **Purpose**: A hands-on lab simulating a full team workflow using Git-driven project management principles on GitHub or a similar platform.
- **Estimated Difficulty**: 4/5
- **Estimated Time**: 90 minutes
- **Prerequisites**: A GitHub account (or GitLab/Bitbucket). All previous lessons in Module 08.

---

### Goal

To simulate a realistic, small-team workflow from issue creation to deployment, using the features of a modern Git hosting platform. This lab is less about specific Git commands and more about the overall process.

### Setup

1.  **Create a Repository**: Go to GitHub (or your preferred platform) and create a new, public repository named `team-workflow-lab`. Initialize it with a `README.md` file.
2.  **Clone the Repository**: Clone the repository to your local machine.
3.  **Create a Project Board**: In the repository's "Projects" tab, create a new project board. Choose the "Kanban" template. This will give you `To do`, `In progress`, and `Done` columns.

### Part 1: The Product Manager - Defining the Work

As the "Product Manager," your job is to define what needs to be done.

1.  **Create an Issue for a New Feature**:
    - Go to the "Issues" tab in your repository.
    - Create a new issue.
    - **Title**: `Feat: Add user greeting to homepage`
    - **Body**:
      ```md
      **As a user, I want to be greeted by name on the homepage so that the experience feels more personal.**

      **Acceptance Criteria:**
      - When a user visits the homepage, it should display the text "Welcome, User!".
      - The greeting should be inside an `<h1>` tag.
      ```
    - **Labels**: Add the labels `feature` and `enhancement`.
    - **Project**: Assign the issue to your `team-workflow-lab` project board. It should appear in the `To do` column.

2.  **Create an Issue for a Bug**:
    - Create another new issue.
    - **Title**: `Bug: README has a typo`
    - **Body**:
      ```md
      **There is a typo in the `README.md` file.**

      **Steps to Reproduce:**
      1. View the `README.md` file.
      2. Observe the typo.

      **Expected Behavior:**
      - The file should be grammatically correct.
      ```
    - **Labels**: Add the label `bug`.
    - **Project**: Assign this issue to your project board as well.

### Part 2: The Developer - Working on the Bug

Now, put on your "Developer" hat. You decide to tackle the easy bug first.

1.  **Assign the Issue**: On the issue page for the typo bug, assign it to yourself.
2.  **Create a Branch**: On your local machine, create a branch for the bug fix. Name it according to the issue number.
    ```bash
    # Let's say the bug issue is #2
    $ git switch -c bug/2-fix-readme-typo
    ```
3.  **Fix the Bug**: Edit the `README.md` file and fix the typo.
4.  **Commit the Fix**: Craft a good commit message.
    ```bash
    $ git add README.md
    $ git commit -m "fix(docs): Correct typo in README

    Closes: #2"
    ```
    The `Closes: #2` footer is important!

5.  **Push the Branch**:
    `$ git push -u origin bug/2-fix-readme-typo`

6.  **Open a Pull Request**:
    - Go to your repository on GitHub. It will likely show a prompt to "Compare & pull request" for the branch you just pushed. Click it.
    - The PR title and body will be pre-filled from your commit message.
    - The PR should automatically be linked to Issue #2.
    - Assign yourself as the reviewer (for this lab).
    - Notice that the issue on your project board may have automatically moved to the `In progress` or `In review` column.

7.  **Merge the PR**:
    - Since it's a simple fix, you review and approve it yourself.
    - Use the "Merge pull request" button.
    - After merging, delete the remote branch.
    - Go back to the "Issues" tab. You should see that Issue #2 is now closed automatically. The card on your project board should be in the `Done` column.

### Part 3: The Developer - Working on the Feature

Now for the more complex feature.

1.  **Assign and Branch**: Assign Issue #1 to yourself and create a local branch.
    ```bash
    $ git switch main
    $ git pull # Make sure you have the fix from the last PR
    $ git switch -c feat/1-user-greeting
    ```
2.  **Implement the Feature**: Create a new `index.html` file.
    ```html
    <!DOCTYPE html>
    <html>
    <head>
      <title>Welcome</title>
    </head>
    <body>
      <h1>Welcome, User!</h1>
    </body>
    </html>
    ```
3.  **Commit and Push**:
    ```bash
    $ git add index.html
    $ git commit -m "feat: Add user greeting to homepage

    Implements the user greeting feature as per the acceptance
    criteria.

    Resolves: #1"
    $ git push -u origin feat/1-user-greeting
    ```
4.  **Open a Pull Request**:
    - Open a PR for this branch.
    - This time, mark it as a "Draft" pull request. This signals to the team that it's a work in progress and not yet ready for a final review.
    - Add a comment to the PR: "@me What do you think of this initial implementation?"

5.  **Add More Changes**: You realize you forgot something.
    - Locally, add a CSS file `style.css` to make it look nicer.
    - Commit and push the change to the same feature branch.
    - Notice that the PR is automatically updated with your new commit.

6.  **Finalize and Merge**:
    - When you're ready, click the "Ready for review" button on the PR.
    - Perform the review, approve, and merge the PR.
    - Check your project board. Everything should now be in the `Done` column.

### Debrief

This lab walked you through a complete, albeit simplified, project management lifecycle driven by Git and GitHub.

- **Visibility**: At any point, anyone on the team could look at the project board or the issues list to see exactly what work was being done, who was doing it, and what its status was.
- **Traceability**: The final `main` branch history contains commits that are directly linked to the PRs they came from, which are in turn linked to the Issues that defined the requirements. You have a full audit trail from idea to deployment.
- **Automation**: You saw how certain keywords (`Closes`, `Resolves`) and actions (opening a PR) can automate the project management process, reducing administrative overhead.

This Git-driven workflow is the foundation of how most modern software teams operate.
