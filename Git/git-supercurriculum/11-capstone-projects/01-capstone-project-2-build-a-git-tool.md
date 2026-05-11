
# 01-capstone-project-2-build-a-git-tool.md

- **Purpose**: To solidify your understanding of Git's internals by building a custom tool that interacts with the Git object model or automates a workflow.
- **Estimated Difficulty**: 5/5
- **Estimated Time**: 1-2 weeks
- **Prerequisites**: All modules, especially `01-git-internals` and `09-tooling-and-productivity`. Strong programming skills in a language of your choice (e.g., Python, Go, Rust, Node.js).

---

### Goal

The goal of this project is to go beyond *using* Git and start *thinking like* Git. You will write a script or application that performs a useful Git-related task by either shelling out to the Git command line or, for a more advanced challenge, by reading from the `.git` directory directly.

This project is open-ended. The key is to choose a small, well-defined problem and solve it.

### Project Ideas (Choose One or Create Your Own)

#### Idea 1: "Git-Stats" - A Repository Statistics Generator

- **Concept**: A command-line tool that you can run in a Git repository to generate interesting statistics.
- **Features**:
    - Total number of commits.
    - Top 5 contributors (by number of commits).
    - A pie chart or bar graph of file types in the repository (e.g., 60% `.js`, 20% `.css`, 10% `.md`).
    - "Code churn" report: Show the number of lines added vs. lines deleted over the last 30 days.
- **Implementation**:
    - **Easy**: Shell out to `git log`, `git ls-files`, etc., and parse the output. For example, `git log --pretty=format:'%an'` will give you a list of all authors, which you can then count and sort.
    - **Hard**: Use a library (like `go-git` for Go or `nodegit` for Node.js) to interact with the repository programmatically.

#### Idea 2: "Git-Doctor" - A Repository Health Checker

- **Concept**: A tool that analyzes a repository for common problems or anti-patterns.
- **Features**:
    - Find large files that should be in Git LFS.
    - Find merged local branches that can be safely deleted.
    - Detect `TODO` or `FIXME` comments and generate a report.
    - Check if the `main` branch has any direct commits that didn't come from a merge (often a sign of a bad workflow).
- **Implementation**:
    - Use `git rev-list --all --objects` and `git cat-file --batch-check` to find large objects.
    - Use `git branch --merged main` to find merged branches.
    - Use `git grep` to find `TODO` comments.
    - Use `git log main --merges` vs. `git log main` to analyze the commit history of `main`.

#### Idea 3: "Git-Recreate" - A Git Internals Lab

- **Concept**: A script that recreates a simple commit from scratch using only plumbing commands, demonstrating your understanding of the object model.
- **Features**:
    - Take a file's content as input.
    - Use `git hash-object` to create a blob.
    - Use `git update-index` to add the blob to the index.
    - Use `git write-tree` to create a tree object from the index.
    - Use `git commit-tree` to create a commit object, pointing to the new tree and the current `HEAD` as its parent.
    - Use `git update-ref` to move the current branch's `HEAD` to the new commit.
- **Implementation**: This is a pure shell scripting project. The goal is to produce a script that, when run, creates a new commit without using `git add` or `git commit`. This is a direct implementation of the lab from Module `01-git-internals`.

### Development and Submission Process

**1. Choose Your Project and Language**
- Pick the idea that sounds most interesting to you.
- Choose a programming language you are comfortable with.

**2. Plan Your Approach**
- Break the problem down into smaller pieces.
- For each piece, determine which Git command(s) will give you the information you need.
- Think about how you will parse the output of those commands.

**3. Build and Iterate**
- Start with the simplest feature and get it working.
- Commit your work frequently! You are building a Git tool, so use Git to do it.
- Write a good `README.md` file that explains what your tool does and how to run it.

**4. "Self-Review"**
- Once your project is "complete," create a pull request for yourself (e.g., from a `develop` branch to `main`) in a repository on GitHub.
- Go through the PR and review your own code as if you were a teammate.
    - Is the code clean and readable?
    - Is the `README` clear?
    - Did you follow good commit practices?
- Merge the PR.

### Debrief

This capstone project moves you from being a Git *user* to a Git *power user* or even a *tool builder*. By forcing you to interact with Git's output and object model programmatically, you will gain a much deeper and more fundamental understanding of how Git works under the hood.

This is the kind of project that demonstrates true mastery. It shows that you don't just know the commands; you understand the system.
