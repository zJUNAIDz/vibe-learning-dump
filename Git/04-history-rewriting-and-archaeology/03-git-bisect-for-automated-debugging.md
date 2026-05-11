
# 03-git-bisect-for-automated-debugging.md

- **Purpose**: To teach `git bisect`, a powerful tool for automatically finding the exact commit that introduced a bug.
- **Estimated Difficulty**: 4/5
- **Estimated Reading Time**: 40 minutes
- **Prerequisites**: `02-finding-fault-with-git-blame-and-git-log.md`

---

### The Problem: The Needle in the Haystack

A bug is reported in production. You know the application was working correctly last week, but it's broken now. In the last week, 500 commits have been merged into `main`. How do you find the one commit that introduced the bug?

You could manually check out commits and test them, but that's slow and tedious. This is the exact problem `git bisect` was designed to solve.

### The Concept: Binary Search for Bugs

`git bisect` performs a binary search on your commit history. You give it a "good" commit (where the bug is not present) and a "bad" commit (where the bug is present). It will then automatically check out a commit halfway between the two and ask you: "Is this commit good or bad?"

Based on your answer, it halves the search space and repeats the process. For 500 commits, you'll only have to test about 9 times (`log2(500) ≈ 9`) to find the exact commit that introduced the bug.

**Diagram: The Bisect Process**
```mermaid
graph TD
    subgraph "Commit History"
        A["Good"] --> B --> C --> D --> E["Bad"]
    end

    subgraph "Step 1"
        Git["Git checks out C (midpoint)"]
        You["You test and say 'git bisect good'"]
    end

    subgraph "Step 2"
        Git2["Search space is now D, E. Git checks out D."]
        You2["You test and say 'git bisect bad'"]
    end

    subgraph "Result"
        Result["Git reports D is the first bad commit"]
    end

    A & E -- "start" --> Git
    Git -- "answer" --> Git2
    Git2 -- "answer" --> Result
```

### The Manual `bisect` Workflow

**1. Start the bisect session**
First, you need to identify a "good" commit and a "bad" commit.
- The "bad" commit is usually `HEAD`.
- The "good" commit might be a tag from last week's release, e.g., `v1.2.0`.

```bash
# Start the bisect process
$ git bisect start

# Tell bisect the current commit is bad
$ git bisect bad HEAD

# Tell bisect the commit from the tag is good
$ git bisect good v1.2.0

# Git will respond:
# Bisecting: 250 revisions left to test after this (roughly 8 steps)
# [a1b2c3d] Some commit message from the midpoint
```

**2. Test and Respond**
Git has now checked out the midpoint commit in a detached HEAD state. Your job is to build the project and test if the bug is present.

- **If the bug exists** in this commit, tell Git:
  `$ git bisect bad`

- **If the bug does not exist** in this commit, tell Git:
  `$ git bisect good`

Git will then check out a new midpoint and prompt you again.

**What if you can't test a commit?** Maybe it doesn't compile or has some other problem. You can tell Git to ignore it and try a different one nearby:
`$ git bisect skip`

**3. The Result**
You repeat the "test and respond" cycle. Eventually, Git will have narrowed it down to a single commit.

```
a9b8c7d is the first bad commit
...
```
Git tells you the exact commit that introduced the bug. You can now inspect it with `git show`, `blame` the author, and write a fix.

**4. End the bisect session**
Once you're done, it's crucial to return to your original branch.

```bash
$ git bisect reset
# You are back at the branch you started on.
```

### The Automated `bisect` Workflow

If you can write a script that can detect the bug, you can automate the entire process. The script must:
- Exit with code `0` if the commit is "good".
- Exit with code `1` if the commit is "bad".
- Exit with code `125` if the commit can't be tested ("skip").

**Example Script (`test-bug.sh`):**
Let's say the bug is that your app's `/health` endpoint returns a 500 error.

```bash
#!/bin/sh

# Rebuild the project
npm install && npm run build

# Start the server in the background
npm start &
SERVER_PID=$!

# Give the server a moment to start
sleep 5

# Check the health endpoint
if curl -s http://localhost:3000/health | grep "OK"; then
  # It's good, exit 0
  kill $SERVER_PID
  exit 0
else
  # It's bad, exit 1
  kill $SERVER_PID
  exit 1
fi
```

Now you can run the whole bisect automatically:
`git bisect run ./test-bug.sh`

Git will run the script on each commit, check its exit code, and do the binary search for you. You can go get a coffee and come back to the answer.

### Key Takeaways

- `git bisect` is a powerful debugging tool that uses a binary search to find the commit that introduced a bug.
- The workflow is `bisect start`, then a loop of `bisect good`/`bad`, and finally `bisect reset`.
- You can automate the entire process with `git bisect run <script>` if you can write a script to detect the bug.
- `bisect` is one of the most impressive and time-saving tools in the Git suite, and mastering it is a mark of a senior developer.

### Interview Notes

- **Question**: "You've discovered a major regression in your `main` branch that was introduced sometime in the last month, spanning hundreds of commits. What's the most efficient way to find the exact commit that caused the problem?"
- **Answer**: "The most efficient tool for this is `git bisect`. I would start by identifying a known 'good' commit from before the bug appeared (like a release tag from last month) and a known 'bad' commit (usually `HEAD`). I'd start the process with `git bisect start`, then `git bisect good <good-commit>` and `git bisect bad <bad-commit>`. Git then performs a binary search on the commit history. For each step, it checks out a commit and I would test for the bug, responding with `git bisect good` or `git bisect bad`. This process narrows down the search space by half each time, allowing me to find the needle in the haystack in a logarithmic number of steps, which is far more efficient than manual checking."
