
# 00-monorepos-vs-polyrepos.md

- **Purpose**: To define and contrast the two major repository structuring strategies, Monorepos and Polyrepos, and discuss their respective trade-offs.
- **Estimated Difficulty**: 3/5
- **Estimated Reading Time**: 30 minutes
- **Prerequisites**: General understanding of repository structure.

---

### What is a Repository Strategy?

As an organization grows, a fundamental architectural decision is how to structure its codebases. Should all projects live in one giant repository, or should every project or service have its own repository? This choice between a "monorepo" and "polyrepo" architecture has profound implications for tooling, collaboration, and development workflows.

### Polyrepo: The Traditional Approach

A polyrepo (or multi-repo) strategy is what most developers are familiar with. Each project, service, or library has its own distinct Git repository.

- **Example**: A company might have separate repositories for its `web-frontend`, `api-server`, `user-service`, `billing-service`, and `shared-ui-library`.

**Pros:**
- **Clear Ownership and Autonomy**: Each repository can have its own team, its own build/test/deploy pipeline, and its own release cadence.
- **Simpler Tooling**: Standard `git` commands work perfectly well. The history of a single project is self-contained and relatively small.
- **Access Control**: It's easy to grant developers access to only the repositories they need to work on.
- **Faster Local Clones**: Cloning a single small repository is fast.

**Cons:**
- **Dependency Management Hell**: The biggest challenge. If `web-frontend` depends on `shared-ui-library`, how do you manage changes? A change in the library requires publishing a new version (e.g., to npm), and then every downstream project needs to update its `package.json` and pull in the new version. This can be slow and complex to coordinate.
- **Difficult to Refactor**: Making a change that spans multiple repositories (e.g., renaming a function in a library and updating all its consumers) is extremely difficult. It requires coordinated pull requests across many repos.
- **Code Discovery and Sharing**: It can be hard to discover code that might be useful from other teams, leading to duplicated effort.

### Monorepo: One Repository to Rule Them All

A monorepo is a strategy where all of an organization's code lives in a single Git repository. This does **not** mean it's a monolith; a monorepo can contain many distinct projects, libraries, and microservices.

- **Example**: Google famously uses a massive monorepo that contains the source code for almost all of its software.

**Pros:**
- **Simplified Dependency Management**: If `web-frontend` and `shared-ui-library` are in the same repository, there is no versioning. A change to the library is immediately available to the frontend. You are always using the "latest" version.
- **Atomic, Cross-Project Commits**: A single commit can change both a library and its consumer. This makes large-scale refactoring trivial. You can change an API and all of its callers in one atomic commit, ensuring the entire system is never in a broken state.
- **Code Sharing and Visibility**: It's easy to search across the entire organization's codebase, promoting code reuse and collaboration.
- **Centralized Tooling**: You can have one build system, one linting configuration, and one set of dependencies for the entire repository.

**Cons:**
- **Tooling Challenges**: Standard `git` can become slow as the repository grows to millions of commits and terabytes of data. This has led to the development of specialized tools like `git sparse-checkout`, the Virtual File System (VFS) for Git, and custom build systems (e.g., Bazel, Buck).
- **Noisy History**: The `main` branch history is a mix of commits from every project in the organization, which can be overwhelming.
- **Broken `main` Breaks Everyone**: A broken build on the `main` branch can block development for the entire company. This requires a very high degree of testing and CI discipline.
- **Large Local Clones**: Cloning the entire repository can be slow and take up a lot of disk space.

### Monorepo Tooling

The challenges of monorepos have led to a rich ecosystem of tools designed to manage them:
- **Build Systems**: Tools like **Bazel** (Google), **Buck** (Facebook), and **Pants** are designed to intelligently build and test only the parts of the repository that were affected by a change.
- **Workspace Managers**: Tools like **Lerna**, **Nx**, and **Turborepo** (for the JavaScript ecosystem) help manage dependencies between projects within the monorepo, run commands across multiple projects, and optimize builds.
- **Git Enhancements**: Features like `sparse-checkout` and `partial-clone` allow developers to check out only the parts of the repository they need to work on, mitigating the large-clone problem.

### Which is Right for You?

| Factor                | Choose Polyrepo if...                               | Choose Monorepo if...                               |
| --------------------- | --------------------------------------------------- | --------------------------------------------------- |
| **Team Size**         | Small to medium.                                    | Any size, but benefits increase with scale.         |
| **Project Coupling**  | Projects are largely independent.                   | Projects are highly interrelated with shared dependencies. |
| **Refactoring Needs** | You rarely need to refactor across project boundaries. | You frequently need to make atomic, cross-project changes. |
| **Tooling Investment**| You want to stick with standard, off-the-shelf tools. | You are willing to invest in specialized build and workspace tools. |
| **Engineering Culture**| Teams are highly autonomous.                        | You want to foster a culture of shared ownership and code reuse. |

**Conclusion**: The polyrepo approach is simpler to start with and is the default for many. The monorepo approach is more complex to set up but solves major pain points around dependency management and cross-project refactoring at scale. The trend in many large tech companies is towards monorepos, enabled by increasingly sophisticated tooling.
