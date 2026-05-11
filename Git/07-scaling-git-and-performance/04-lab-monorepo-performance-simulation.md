
# 04-lab-monorepo-performance-simulation.md

- **Purpose**: A hands-on lab to simulate working in a large monorepo and using `sparse-checkout` and `partial clone` to manage performance.
- **Estimated Difficulty**: 5/5
- **Estimated Time**: 90 minutes
- **Prerequisites**: All previous lessons in Module 07.

---

### Goal

To experience the performance problems of a large monorepo and to use modern Git features to make the developer experience fast and efficient.

### Setup: Create a Simulated Monorepo

We'll create a bare "remote" repository and then script the creation of a large number of files and a moderately deep history to simulate a monorepo.

**1. Create the Bare Remote**
```bash
$ mkdir monorepo-lab && cd monorepo-lab
$ git init --bare remote.git
```

**2. Create the Monorepo Structure**
Now, we'll clone it and add a lot of files.

```bash
$ git clone remote.git initial-setup
$ cd initial-setup

# Create a script to generate files. Create a file named 'generate.sh'
$ cat << 'EOF' > generate.sh
#!/bin/bash
for i in {1..5}; do
  mkdir -p "services/service-$i"
  mkdir -p "libs/lib-$i"
  for j in {1..100}; do
    echo "Service $i, file $j" > "services/service-$i/file-$j.txt"
    echo "Lib $i, file $j" > "libs/lib-$i/file-$j.txt"
  done
done
git add .
git commit -m "Initial commit of services and libs"
EOF

$ chmod +x generate.sh
$ ./generate.sh
$ git push origin main
```
We now have a repository with 10 directories and 1000 files.

**3. Bloat the History**
Let's add more commits to simulate a history.

```bash
# In the 'initial-setup' directory
$ for i in {1..50}; do
  echo "Update $i" >> services/service-1/file-1.txt
  git commit -am "Update service 1 ($i)"
done
$ git push origin main
$ cd ..
```
Our remote monorepo is now ready.

### Part 1: The "Bad" Experience (Full Clone)

First, let's experience the slowness of a full clone.

1.  **Clone the repository normally.**
    ```bash
    $ git clone remote.git full-clone
    $ cd full-clone
    ```
2.  **Check the size and file count.**
    ```bash
    $ ls -lR | wc -l # Count the files
    $ du -sh . # Check disk usage
    ```
3.  **Time `git status`.**
    ```bash
    $ time git status
    # On branch main... etc.
    # real 0m0.XXXs  <- Note this time
    ```
    Even with only 1000 files, you may notice a small but measurable delay. Imagine this with a million files.

### Part 2: The "Good" Experience (`sparse-checkout`)

Now, let's simulate the workflow of a developer who only needs to work on `service-2` and `lib-2`.

1.  **Perform a partial, no-checkout clone.**
    ```bash
    $ cd .. # Go back to the lab root
    $ git clone --filter=blob:none --no-checkout remote.git sparse-clone
    $ cd sparse-clone
    ```
    This should be almost instantaneous, as it's only downloading the commit and tree data, not the file contents.

2.  **Initialize sparse-checkout.**
    You only care about `service-2` and `lib-2`.
    ```bash
    $ git sparse-checkout init --cone
    $ git sparse-checkout set services/service-2 libs/lib-2
    ```

3.  **Check out the `main` branch.**
    ```bash
    $ git checkout main
    ```
    Git will now download the blobs for the files in the directories you specified and populate your working directory.

4.  **Check the size and file count.**
    ```bash
    $ ls -lR | wc -l # Should be around 200 files, not 1000
    $ du -sh . # Should be much smaller than the full clone
    ```

5.  **Time `git status`.**
    ```bash
    $ time git status
    # real 0m0.YYYs <- Note this time
    ```
    The time should be significantly less than the full clone, as Git has far fewer files to scan.

### Part 3: Working with the Sparse Checkout

Let's prove that this is a fully functional Git repository.

1.  **Make a change.**
    ```bash
    $ echo "My new feature" >> services/service-2/file-50.txt
    $ git commit -am "Feat: Add new feature to service 2"
    ```
2.  **Push the change.**
    ```bash
    $ git push origin main
    ```
3.  **Verify the change in the full clone.**
    ```bash
    $ cd ../full-clone
    $ git pull
    $ cat services/service-2/file-50.txt
    # You should see "My new feature" appended.
    ```
This demonstrates that you can work efficiently in a small, sparse view of the repository while still collaborating with others who might have a different view (or a full checkout).

### Debrief

This lab simulated the core problem of monorepos: working directory size.
- The **full clone** was larger and slower to `status`.
- The **sparse clone** workflow (`clone --no-checkout`, `sparse-checkout set`, `checkout`) resulted in a much smaller, faster local developer experience.

By using `sparse-checkout`, you get the main benefits of a monorepo (atomic commits, unified history) without the primary drawback (a slow and bloated local checkout). This is the key technique for making monorepos practical at scale.
