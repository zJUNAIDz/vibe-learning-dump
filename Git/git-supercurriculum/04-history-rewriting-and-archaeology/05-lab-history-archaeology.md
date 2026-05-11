
# 05-lab-history-archaeology.md

- **Purpose**: A hands-on lab where you act as a code detective, using `log`, `blame`, and `bisect` to investigate a bug and its history.
- **Estimated Difficulty**: 5/5
- **Estimated Time**: 90 minutes
- **Prerequisites**: All previous lessons in Module 04.

---

### Goal

To use Git's history and archaeology tools to pinpoint the exact cause and origin of a regression bug in a simulated project.

### The Scenario

You are a developer on a project that calculates shipping costs. A feature was added to give "premium" users a discount. A bug has been reported: sometimes, users are getting a *negative* shipping cost, meaning you are paying them to ship items. This is bad.

You know the app was working correctly at tag `v1.0`. The bug is present at `HEAD`. Your task is to find the commit that introduced this bug.

### Setup: Create the Project History

We will create a small repository with a scripted history to simulate this scenario.

```bash
$ mkdir archaeology-lab && cd archaeology-lab
$ git init

# Initial commit
$ echo 10 > cost_calculator.js
$ git add . && git commit -m "Initial commit with base cost"
$ git tag v1.0

# Add some unrelated work
$ echo "config=true" > config.cfg
$ git add . && git commit -m "Feat: Add configuration"

# Add the premium user feature (correctly)
$ echo 'const cost = 10; const isPremium = true; const finalCost = isPremium ? cost * 0.8 : cost;' > cost_calculator.js
$ git add . && git commit -m "Feat: Add premium user discount"

# More unrelated work
$ echo "Docs" > README.md
$ git add . && git commit -m "Docs: Add README file"

# Introduce the bug! A bad refactor.
$ echo 'const cost = 10; const isPremium = true; const discount = 0.8; const finalCost = isPremium ? cost - discount : cost;' > cost_calculator.js
$ git add . && git commit -m "Refactor: Use discount variable"

# Even more unrelated work
$ echo "More docs" >> README.md
$ git add . && git commit -m "Docs: Update README"
```
The bug is in the "Refactor" commit. The logic changed from multiplication (`cost * 0.8`) to subtraction (`cost - discount`), causing the final cost to be `10 - 0.8 = 9.2` instead of `8`. If the cost were less than the discount, it would be negative.

### Part 1: Initial Investigation (`blame`)

You've identified the file `cost_calculator.js` as the source of the problem.

1.  **Run `git blame` on the file.**
    ```bash
    $ git blame cost_calculator.js
    # You will see the SHAs and authors for each line.
    # The line with the buggy calculation will point to the "Refactor" commit.
    ```
2.  **Inspect the "Refactor" commit.**
    ```bash
    $ git show <sha_of_refactor_commit>
    # The diff will clearly show the change from '*' to '-'.
    ```
In this simple case, `blame` leads us directly to the answer. But what if the history were much more complex, with many changes to that line?

### Part 2: Automated Investigation (`bisect`)

Let's assume the history is 1000 commits long and `blame` isn't so obvious. We'll use `bisect` to find the bug.

1.  **Create a test script.** This script will act as our automated bug detector.
    Create a file named `test.sh`:
    ```bash
    #!/bin/sh

    # Use 'node' to execute the JS file and get the final cost.
    # We add 'console.log(finalCost)' to the end to print the value.
    COST=$(node -e "$(cat cost_calculator.js; echo 'console.log(finalCost);')")

    # Check if the cost is less than 5 (a simple proxy for a negative or too-low cost)
    if [ $(echo "$COST < 5" | bc) -eq 1 ]; then
      exit 1 # It's bad
    else
      exit 0 # It's good
    fi
    ```
    Make the script executable: `chmod +x test.sh`.
    *(Note: This requires `node` and `bc` to be installed on your system).*

2.  **Start the bisect process.**
    ```bash
    $ git bisect start
    ```
3.  **Mark the good and bad commits.**
    ```bash
    $ git bisect bad HEAD
    $ git bisect good v1.0
    ```
4.  **Run the automated bisect.**
    ```bash
    $ git bisect run ./test.sh
    ```
5.  **Analyze the output.** Git will run the script on several commits, performing its binary search. In the end, it will print:
    ```
    <sha_of_refactor_commit> is the first bad commit
    ...
    ```
    `bisect` has automatically found the exact commit that introduced the regression, without you needing to manually test anything.

6.  **Clean up.**
    ```bash
    $ git bisect reset
    ```

### Part 3: Historical Search (`log`)

What if you didn't know *what* the bug was, just that something was wrong with discounts? You could use `log`'s pickaxe to find when the word "discount" was introduced.

1.  **Search the history for a change in a specific string.**
    ```bash
    $ git log -S "discount"
    ```
2.  **Analyze the output.** This command will return only the commit(s) that introduced or removed the word "discount". In our case, it will point directly to the "Refactor" commit. This is an incredibly powerful way to find the origin of a specific piece of code.

### Debrief

This lab demonstrated three different approaches to code archaeology:
- `git blame`: Good for finding the last person to touch a line, which is often a great starting point.
- `git bisect`: The ultimate tool for finding a regression in a long history. Automating it with a script is a massive time-saver.
- `git log -S`: A "pickaxe" for finding the exact moment a specific string or variable was introduced or removed, perfect for tracking down the origin of a concept in the codebase.

Mastering these three tools will make you an effective code detective, able to diagnose the history of any bug.
