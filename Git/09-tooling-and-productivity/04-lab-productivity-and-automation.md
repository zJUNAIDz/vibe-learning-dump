
# 04-lab-productivity-and-automation.md

- **Purpose**: A hands-on lab to set up a productive local Git environment with aliases and hooks, and to create a basic CI pipeline with GitHub Actions.
- **Estimated Difficulty**: 4/5
- **Estimated Time**: 90 minutes
- **Prerequisites**: All previous lessons in Module 09. A GitHub account. Node.js and npm installed locally.

---

### Goal

To configure a repository from scratch with tools that improve productivity and automate quality checks, simulating a professional project setup.

### Part 1: Local Productivity Setup

**1. Create and Clone the Repository**
- On GitHub, create a new repository named `productivity-lab`.
- Initialize it with a `README.md` and a `.gitignore` for `Node`.
- Clone it to your local machine.

**2. Set Up a Node.js Project**
We need a project to work with.
```bash
$ cd productivity-lab
$ npm init -y
$ npm install --save-dev eslint prettier
```

**3. Configure Aliases**
Let's add some of the useful aliases from the lesson to your **global** Git config.

```bash
$ git config --global alias.st status
$ git config --global alias.co checkout
$ git config --global alias.br branch
$ git config --global alias.lg "log --graph --oneline --decorate"
$ git config --global alias.unstage "restore --staged"
```
Test them out: `git st`, `git lg`.

**4. Set Up a Pre-commit Hook with Husky**
We want to automatically run `prettier` on our code before we commit it. We'll use Husky to manage the hook.

- **Install Husky**:
  `$ npx husky-init && npm install`
  This will create a `.husky` directory and a sample `pre-commit` hook.

- **Configure the hook**: Modify the `.husky/pre-commit` file to run Prettier.
  ```bash
  #!/bin/sh
  . "$(dirname "$0")/_/husky.sh"

  npx prettier --write .
  git add .
  ```
  This script will format all files and then stage the changes from the formatting.

- **Test the hook**:
  - Create a new file, `index.js`, with intentionally bad formatting.
    ```javascript
    const x =   1;
    function foo( ) {
    return x
    }
    ```
  - Stage the file: `git add index.js`.
  - Commit it: `git commit -m "Add index file"`.
  - **Observe**: Before the commit is created, Husky will run the `pre-commit` hook. Prettier will reformat the file. The hook will then `git add` the reformatted file, so your commit will contain the clean version.
  - Check the contents of `index.js` and the `git log` to confirm.

### Part 2: Automation with GitHub Actions

Now, let's create a CI pipeline to automatically run a linter and tests.

**1. Configure ESLint**
- Set up a basic ESLint configuration.
  `$ npx eslint --init`
  (Follow the prompts; choose "To check syntax and find problems," "JavaScript modules," "None of these," "No," "JavaScript," and answer yes to install dependencies).
- Add a `lint` script to your `package.json`:
  ```json
  "scripts": {
    "lint": "eslint ."
  },
  ```

**2. Create the CI Workflow File**
- Create the directory structure: `.github/workflows`.
- Create a new file: `.github/workflows/ci.yml`.
- Add the following content:
  ```yaml
  name: CI

  on:
    push:
      branches: [ main ]
    pull_request:
      branches: [ main ]

  jobs:
    lint:
      runs-on: ubuntu-latest
      steps:
        - name: Checkout code
          uses: actions/checkout@v3
        - name: Setup Node.js
          uses: actions/setup-node@v3
          with:
            node-version: '18'
            cache: 'npm'
        - name: Install dependencies
          run: npm ci
        - name: Run linter
          run: npm run lint
  ```

**3. Commit and Push**
Commit all the changes you've made (the Husky setup, the `ci.yml` file, etc.).
`$ git add .`
`$ git commit -m "chore: Configure productivity and CI tools"`
`$ git push origin main`

**4. Test the CI Pipeline**
- Go to the "Actions" tab of your repository on GitHub. You should see your "CI" workflow running and, hopefully, passing.

**5. Test the Pull Request Integration**
- Create a new branch: `git switch -c feature/test-ci`.
- Introduce a linting error in `index.js`. For example, declare a variable but don't use it.
  ```javascript
  const y = 2; // Unused variable
  ```
- Commit and push this branch.
  `$ git commit -am "feat: Add unused variable to test CI"`
  `$ git push -u origin feature/test-ci`
- **Open a Pull Request** on GitHub for this branch, targeting `main`.
- **Observe**: The "CI" workflow will automatically start running on the PR. After a minute, it should fail. The PR page will show a red "X" next to the check, indicating that it's not safe to merge.
- **Configure Branch Protection**:
  - Go to Settings -> Branches in your repository.
  - Add a branch protection rule for `main`.
  - Check the box for "Require status checks to pass before merging."
  - Select the "lint" job from your CI workflow.
  - Save the rule.
- Now, go back to your PR. You should see a message that the branch is protected and cannot be merged until the checks pass.

### Debrief

This lab has given you a taste of a professional developer's setup.
- **Locally**, you have aliases for speed and a `pre-commit` hook that automatically formats your code, ensuring consistency and saving you time.
- **Remotely**, you have a GitHub Actions workflow that acts as a gatekeeper for your `main` branch. It automatically verifies the quality of every push and pull request, preventing broken code from being merged.

This combination of local productivity hacks and remote automation is the key to maintaining a high-quality, fast-moving software project.
