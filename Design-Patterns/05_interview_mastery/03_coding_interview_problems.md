
# Interview Mastery: Coding Interview Problems

Design patterns can be a powerful tool for solving coding interview problems, especially those that involve object-oriented modeling or complex interactions between components. Recognizing that a problem can be solved with a known pattern can save you a lot of time and lead to a much cleaner, more extensible solution.

Here are a few examples of common coding interview problems and how they map to design patterns.

---

## 1. Problem: Implement an In-Memory File System

*   **The Prompt:** Design and implement a data structure that represents a file system. It should support operations like `mkdir` (make directory), `addContentToFile`, and `readContentFromFile`. The `ls` command should list the contents of a directory.

*   **The Pattern: Composite Pattern**
    *   This is the canonical example of the Composite pattern. The file system is a tree structure where some nodes are "leaf" nodes (Files) and others are "branch" nodes (Directories). You want to be able to treat both of them uniformly.

*   **The Solution:**
    1.  Create a common `FileSystemNode` interface or abstract class with methods like `getName()` and `getType()`.
    2.  Create a `File` class that implements `FileSystemNode`. It will have properties for its content.
    3.  Create a `Directory` class that also implements `FileSystemNode`. It will contain a list of other `FileSystemNode` objects (`children`).
    4.  The `Directory` class will have methods like `addChild()` and `getChildren()`.
    5.  The main `FileSystem` class will manage the root directory and orchestrate the operations.

*   **TypeScript Implementation:**

    ```typescript
    interface FileSystemNode {
      name: string;
      isDirectory(): boolean;
    }

    class File implements FileSystemNode {
      public content: string = '';
      constructor(public name: string) {}
      isDirectory(): boolean { return false; }
    }

    class Directory implements FileSystemNode {
      public children: Map<string, FileSystemNode> = new Map();
      constructor(public name: string) {}
      isDirectory(): boolean { return true; }
    }

    class FileSystem {
      root: Directory = new Directory('/');

      // Example: ls
      ls(path: string): string[] {
        const node = this.findNode(path);
        if (node && node.isDirectory()) {
          return Array.from((node as Directory).children.keys());
        }
        if (node) { // It's a file
          return [node.name];
        }
        return [];
      }
      // ... other methods like mkdir, addContentToFile
    }
    ```

---

## 2. Problem: Design a Vending Machine

*   **The Prompt:** Design a vending machine that accepts coins, allows a user to select an item, dispenses the item, and provides change. It should handle states like "no coins inserted," "coins inserted," "item sold," and "out of stock."

*   **The Pattern: State Pattern**
    *   The machine's behavior changes drastically based on its current state. This is a classic state machine problem, and the State pattern is the object-oriented way to solve it.

*   **The Solution:**
    1.  Create a `VendingMachine` class (the `Context`). It will hold the inventory and a reference to the current state.
    2.  Create a `State` interface with methods for all possible actions: `insertCoin()`, `selectItem()`, `dispenseItem()`.
    3.  Create concrete state classes: `NoCoinState`, `HasCoinState`, `SoldState`, `EmptyState`.
    4.  Each state class implements the interface. For example, in `NoCoinState`, `insertCoin()` will transition the machine to `HasCoinState`. `selectItem()` will do nothing.
    5.  The `VendingMachine` delegates all actions to its current state object.

*   **TypeScript Implementation:**

    ```typescript
    class VendingMachine { // Context
      public currentState: State;
      public inventory: number;
      constructor() {
        this.inventory = 10;
        this.currentState = new NoCoinState(this);
      }
      // ... delegates methods to currentState
    }

    interface State {
      insertCoin(): void;
      selectItem(): void;
    }

    class NoCoinState implements State { // Concrete State
      constructor(private machine: VendingMachine) {}
      insertCoin() {
        console.log('Coin inserted.');
        this.machine.currentState = new HasCoinState(this.machine);
      }
      selectItem() {
        console.log('Please insert a coin first.');
      }
    }
    // ... other states (HasCoinState, etc.)
    ```

---

## 3. Problem: Implement an Undo/Redo Feature for a Text Editor

*   **The Prompt:** Design a simple text editor that supports typing characters and an `undo()` and `redo()` functionality.

*   **The Patterns: Command and Memento**
    *   This problem can be solved in two excellent ways, and discussing the trade-offs is a great way to impress an interviewer.

*   **Solution 1 (Command Pattern):**
    1.  Treat every user action as a command. Create a `TypeCharacterCommand` that knows which character was typed and at what position.
    2.  The `execute()` method of the command inserts the character. The `undo()` method removes it.
    3.  Maintain two stacks: an `undoStack` and a `redoStack`.
    4.  When a command is executed, push it onto the `undoStack`.
    5.  When `undo()` is called, pop from the `undoStack`, call the command's `undo()` method, and push the command onto the `redoStack`.
    6.  When `redo()` is called, pop from the `redoStack`, call its `execute()` method, and push it back onto the `undoStack`.

*   **Solution 2 (Memento Pattern):**
    1.  The `TextEditor` is the `Originator`.
    2.  Create a `Memento` object that can store the entire content of the editor as a string.
    3.  The `History` class is the `Caretaker`. It has an `undoStack` and a `redoStack` of mementos.
    4.  Before any change, save a memento of the current state to the `undoStack`.
    5.  When `undo()` is called, pop a memento and restore the editor's state from it.

*   **Discussion of Trade-offs:**
    *   **Command** is more memory-efficient if the state is large and the actions are small. You only store the small change, not the entire document.
    *   **Memento** is simpler to implement if the state is small (like a string). It's also more robust, as you're restoring to a known good state rather than trying to reverse an operation, which can sometimes be complex.

---

## 4. Problem: Design a Caching System (LRU Cache)

*   **The Prompt:** Design a Least Recently Used (LRU) cache. It should support `get(key)` and `put(key, value)` operations. When the cache is full and a new item is added, the least recently used item should be evicted.

*   **The Pattern: Proxy (and others)**
    *   While this is more of a data structure problem, you can frame it within design patterns. The cache itself can be seen as a **Proxy** or **Decorator** for a slower data source (like a database or a network call).

*   **The Solution:**
    1.  The core of the problem is implementing the LRU logic efficiently. The classic solution is to use a combination of a **Hash Map** (for O(1) lookups) and a **Doubly Linked List** (for O(1) additions/removals of the most/least recent items).
    2.  The `Cache` class acts as a `Proxy`. When `get(key)` is called:
        *   It first checks its internal storage (the hash map).
        *   If the item exists (a "cache hit"), it moves the item to the front of the linked list and returns it.
        *   If the item doesn't exist (a "cache miss"), it fetches the data from the original source (the "real subject"), stores it in the cache (and evicts an old item if necessary), and then returns it.

*   **TypeScript Implementation:**

    ```typescript
    class LruCache {
      private capacity: number;
      private cache: Map<any, any> = new Map();
      private realService: RealService; // The object being proxied

      constructor(capacity: number, realService: RealService) {
        this.capacity = capacity;
        this.realService = realService;
      }

      async get(key: any): Promise<any> {
        if (this.cache.has(key)) {
          // Cache hit logic (update recency)
          return this.cache.get(key);
        } else {
          // Cache miss
          const value = await this.realService.fetch(key);
          this.put(key, value);
          return value;
        }
      }
      // ... put and eviction logic
    }
    ```
