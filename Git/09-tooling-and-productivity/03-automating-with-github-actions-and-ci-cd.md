
# 03-automating-with-github-actions-and-ci-cd.md

- **Purpose**: To provide an introduction to CI/CD principles and demonstrate how to automate workflows using GitHub Actions.
- **Estimated Difficulty**: 4/5
- **Estimated Reading Time**: 50 minutes
- **Prerequisites**: Basic understanding of YAML syntax. `00-git-hooks-an-introduction.md`.

---

### What is CI/CD?

CI/CD is a cornerstone of modern software development and DevOps.

- **Continuous Integration (CI)**: The practice of frequently merging all developers' working copies to a shared mainline (e.g., the `main` or `develop` branch). Each integration is then automatically verified by a build and a set of automated tests. The goal is to detect integration issues as early as possible.

- **Continuous Deployment (or Delivery) (CD)**: The practice of automatically deploying every change that passes the CI stage to a testing or production environment. This allows for rapid, reliable delivery of new features and fixes.

Git is the foundation of CI/CD. The entire process is triggered by Git events like `push` and `pull_request`.

### GitHub Actions: CI/CD Built into GitHub

GitHub Actions is a powerful and flexible CI/CD platform built directly into GitHub. It allows you to define automated workflows that respond to any event in your repository.

- **Workflows** are defined in YAML files stored in the `.github/workflows` directory of your repository.
- A repository can have multiple workflow files.
- A **workflow** is made up of one or more **jobs**.
- A **job** is a set of **steps** that execute on a virtual machine (a "runner").
- A **step** can be a shell command or a reusable **action** (a pre-packaged script).

### Workflow 1: A Basic CI Pipeline

Let's create a workflow that runs on every push and pull request to the `main` branch. It will check out the code, install dependencies, run a linter, and run tests.

Create the file `.github/workflows/ci.yml`:

```yaml
name: CI Pipeline

# 1. Trigger: When does this workflow run?
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

# 2. Jobs: What work should be done?
jobs:
  build-and-test:
    # 3. Runner: What kind of machine should this run on?
    runs-on: ubuntu-latest

    # 4. Steps: What are the individual commands?
    steps:
      # Step 1: Check out the repository's code
      - name: Checkout code
        uses: actions/checkout@v3

      # Step 2: Set up the Node.js environment
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm' # Cache dependencies for faster builds

      # Step 3: Install dependencies
      - name: Install dependencies
        run: npm ci

      # Step 4: Run the linter
      - name: Run linter
        run: npm run lint

      # Step 5: Run the tests
      - name: Run tests
        run: npm test
```

**How it works:**
1.  When you push to `main` or open a PR against `main`, GitHub automatically triggers this workflow.
2.  It spins up a fresh Ubuntu virtual machine.
3.  It runs through the steps in order.
    - `actions/checkout@v3` and `actions/setup-node@v3` are reusable actions provided by GitHub and the community.
    - `run:` executes a shell command.
4.  If any step fails (exits with a non-zero code), the entire job fails.
5.  On a pull request, this job's status (pass/fail) will be displayed directly on the PR page. You can configure a branch protection rule to block merging if the CI checks fail.

### Workflow 2: A Simple CD Pipeline

Let's extend the previous example. If the CI pipeline passes on a push to `main`, we want to automatically deploy the code to a hosting service (e.g., GitHub Pages, Vercel, AWS).

Create the file `.github/workflows/cd.yml`:

```yaml
name: CD Pipeline

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest

    # This job depends on the successful completion of the CI job from the other file.
    # (This is a simplified example; a real one might have the build as a separate job here)
    
    steps:
      - name: Checkout code
        uses: actions/checkout@v3

      # ... steps to build the application ...
      - name: Build project
        run: npm run build

      # This step would be specific to your hosting provider.
      # Many providers have dedicated GitHub Actions for deployment.
      - name's: Deploy to Vercel
        uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }} # Use encrypted secrets for tokens
          vercel-org-id: ${{ secrets.ORG_ID }}
          vercel-project-id: ${{ secrets.PROJECT_ID }}
          vercel-args: '--prod'
```

**Key Concepts:**
- **Secrets**: The `VERCEL_TOKEN` is a secret API key. You should never paste secrets directly into your YAML files. Instead, you store them in your repository's Settings -> Secrets and variables -> Actions. GitHub Actions makes them available as environment variables.
- **Marketplace**: There is a huge marketplace of pre-built actions for almost any task (deploying to any cloud, sending a Slack notification, etc.). You should always look for an existing action before writing your own script.

### The Power of Automation

By combining Git with a CI/CD platform like GitHub Actions, you can automate the entire development lifecycle:
- **Enforce Quality**: Automatically run linters and tests on every change, preventing bugs from ever reaching the main branch.
- **Improve Velocity**: Automate the deployment process, allowing you to ship features and fixes to users in minutes, not days.
- **Increase Confidence**: When you have a robust CI/CD pipeline, you can merge and deploy with confidence, knowing that a suite of automated checks has been run.
- **Git-Driven Workflow**: The state of your Git repository *is* the state of your application. Merging to `main` means "release this."

### Key Takeaways

- **CI (Continuous Integration)** is about merging and testing code frequently.
- **CD (Continuous Deployment)** is about automatically deploying code that passes CI.
- **GitHub Actions** is a powerful platform for creating CI/CD workflows that are triggered by Git events.
- Workflows are YAML files in the `.github/workflows` directory.
- Use actions from the marketplace to perform common tasks like checking out code and deploying to cloud providers.
- Use encrypted secrets to store sensitive information like API tokens.
- A good CI/CD pipeline is a critical part of modern, professional software development.
