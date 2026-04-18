
# 🗿 SOLID Principles: The Deep Dive Your Manager Never Gave You

You've probably seen the SOLID acronym in a blog post or had it thrown at you in an interview. Most developers can name the principles, but few can explain what they *actually* mean in a practical, non-bullshit way.

Let's fix that.

SOLID isn't a set of rigid laws. It's a collection of design principles that, when applied correctly, help you avoid creating a tangled mess of code. They are guide rails to keep your OOP design from driving off a cliff.

---

## S: Single Responsibility Principle (SRP)

*   **The Lie:** "A class should do only one thing."
*   **The Problem:** "One thing" is hopelessly vague. Is `Drone.move()` one thing? Or is it calculating vectors, updating coordinates, and draining the battery? That's three things! This definition is useless.
*   **The Truth (from Uncle Bob):** "A class should have only one reason to change."

This is far more powerful. It's not about the number of methods, but about the *actors* and *concerns* that might force a change.

**Example of a Violation:**

Imagine a `Report` class.

```typescript
class Report {
  generateReport(data: any[]): string {
    // 1. Logic to format the data into a string (e.g., CSV)
    let report = "ID,Name,Value\n";
    for (const item of data) {
      report += `${item.id},${item.name},${item.value}\n`;
    }
    return report;
  }

  saveToFile(report: string, filePath: string) {
    // 2. Logic to write the string to the file system
    const fs = require('fs');
    fs.writeFileSync(filePath, report);
  }
}
```

This class has **two reasons to change**:
1.  **The C-level execs change their mind about the report format.** They want JSON now, not CSV. You have to change `generateReport`.
2.  **The DevOps team decides to move from local disk to S3.** You have to change `saveToFile`.

These two concerns—formatting and persistence—are unrelated. They should not live in the same class.

**A Better Approach (SRP Compliant):**

```typescript
// Concern 1: Formatting the report
class ReportFormatter {
  formatAsCsv(data: any[]): string {
    // ... formatting logic
  }
}

// Concern 2: Saving the report
interface ReportPersistence {
  save(report: string, destination: string): void;
}

class FileSystemPersistence implements ReportPersistence {
  save(report: string, filePath: string) {
    // ... fs.writeFileSync logic
  }
}

class S3Persistence implements ReportPersistence {
  save(report: string, bucketUrl: string) {
    // ... S3 SDK logic
  }
}
```

Now, if the report format changes, you only touch `ReportFormatter`. If the storage mechanism changes, you only touch the persistence classes. Each class has a single, clear reason to change.

---

## O: Open/Closed Principle (OCP)

*   **The Lie:** "You shouldn't change existing code."
*   **The Problem:** This is obviously impossible. You have to change code to fix bugs or add features.
*   **The Truth:** "Software entities (classes, modules, functions) should be open for extension, but closed for modification."

This means you should be able to **add new functionality without changing existing, working code.** The primary mechanism for this is **abstraction**.

**Example of a Violation:**

Imagine a `PaymentProcessor` that handles credit card payments.

```typescript
class PaymentProcessor {
  processPayment(amount: number, type: 'credit' | 'paypal') {
    if (type === 'credit') {
      // Logic for credit card API
      console.log(`Processing credit card payment of $${amount}`);
    } else if (type === 'paypal') {
      // Logic for PayPal API
      console.log(`Processing PayPal payment of $${amount}`);
    }
    // ... what happens when we add Apple Pay? Or Crypto?
  }
}
```

To add a new payment method, you have to go back and **modify** the `PaymentProcessor` class. You have to add another `else if`. This is a violation. Every new payment type risks breaking the logic for the existing ones.

**A Better Approach (OCP Compliant):**

Use a common interface and let new payment methods be **extensions**.

```typescript
// The Abstraction (closed for modification)
interface PaymentProvider {
  process(amount: number): void;
}

// The Extensions (open for extension)
class CreditCardProvider implements PaymentProvider {
  process(amount: number) {
    console.log(`Processing credit card payment of $${amount}`);
  }
}

class PayPalProvider implements PaymentProvider {
  process(amount: number) {
    console.log(`Processing PayPal payment of $${amount}`);
  }
}

// Now, let's add a new one without touching anything above
class CryptoProvider implements PaymentProvider {
  process(amount: number) {
    console.log(`Processing crypto payment of $${amount}`);
  }
}

// The client code doesn't change
class PaymentProcessor {
  processPayment(amount: number, provider: PaymentProvider) {
    provider.process(amount);
  }
}

const processor = new PaymentProcessor();
processor.processPayment(100, new CreditCardProvider());
processor.processPayment(50, new CryptoProvider()); // Added without modification!
```

---

## L: Liskov Substitution Principle (LSP)

*   **The Lie:** "A child class should be able to substitute for its parent."
*   **The Problem:** Too generic. What does "substitute" mean?
*   **The Truth:** "Subtypes must be substitutable for their base types **without altering the correctness of the program**."

This is the principle we violated in the `Bird`/`Penguin` example. If you have code that works with a `Bird`, it should also work with a `Penguin` without blowing up or producing weird results.

**The Classic `Rectangle`/`Square` Violation:**

This one is famous for a reason. Mathematically, a square *is a* rectangle. So, let's model that with inheritance.

```typescript
class Rectangle {
  protected width: number;
  protected height: number;

  setWidth(width: number) { this.width = width; }
  setHeight(height: number) { this.height = height; }

  getArea(): number {
    return this.width * this.height;
  }
}

class Square extends Rectangle {
  // A square's width and height must be the same.
  // So we have to override the setters to enforce this rule.
  setWidth(width: number) {
    this.width = width;
    this.height = width; // This is the problem
  }
  setHeight(height: number) {
    this.width = height;
    this.height = height; // This is the problem
  }
}

// A function that works perfectly fine with the base class
function testArea(rect: Rectangle) {
  rect.setWidth(5);
  rect.setHeight(4);
  const area = rect.getArea();

  // The programmer reasonably expects area to be 20.
  console.log(`Expected 20, got ${area}`);
}

const rect = new Rectangle();
const sq = new Square();

testArea(rect); // "Expected 20, got 20" -> Correct!
testArea(sq);   // "Expected 20, got 16" -> WTF?!
```

The `Square` class, by changing the behavior of the parent's methods in an unexpected way, breaks the `testArea` function. It is **not** a substitutable subtype. This is a classic sign that inheritance was the wrong tool for the job.

---

## I: Interface Segregation Principle (ISP)

*   **The Lie:** "Don't make big interfaces."
*   **The Problem:** "Big" is subjective.
*   **The Truth:** "Clients should not be forced to depend on methods they do not use."

This is about making your interfaces focused and role-based. If a class implements an interface, it should need *every single method* on that interface.

**Example of a Violation:**

Imagine a "fat" interface for managing documents.

```typescript
// A "God" interface
interface DocumentManager {
  open(path: string): void;
  close(): void;
  save(): void;
  print(): void;
  exportToPdf(): void;
  sendByEmail(recipient: string): void;
}

// This class only cares about reading and printing.
// But it's forced to implement methods it doesn't need.
class ReadOnlyPrinter implements DocumentManager {
  open(path: string) { /* ... */ }
  close() { /* ... */ }
  print() { /* ... */ }

  // Useless, forced implementations
  save() {
    throw new Error('This is a read-only printer, cannot save.');
  }
  exportToPdf() {
    throw new Error('This is a read-only printer, cannot export.');
  }
  sendByEmail(recipient: string) {
    throw new Error('This is a read-only printer, cannot send email.');
  }
}
```

This is a trap. The `ReadOnlyPrinter` is now a landmine waiting to explode if someone calls `save()`.

**A Better Approach (ISP Compliant):**

Break the fat interface down into smaller, role-based interfaces.

```typescript
interface Readable {
  open(path: string): void;
  close(): void;
}

interface Writable {
  save(): void;
}

interface Printable {
  print(): void;
}

interface Shareable {
  exportToPdf(): void;
  sendByEmail(recipient: string): void;
}

// Now, classes implement only what they need.
class FullEditor implements Readable, Writable, Printable, Shareable {
  // ... implements all methods
}

class ReadOnlyPrinter implements Readable, Printable {
  // ... only needs to implement open, close, and print.
  // No more landmines!
}
```

Clients can now depend on the smallest possible interface they need, leading to much more decoupled and maintainable code.

---

## D: Dependency Inversion Principle (DIP)

*   **The Lie:** "Depend on interfaces, not classes."
*   **The Problem:** This is part of it, but it misses the "inversion" part.
*   **The Truth:**
    1.  High-level modules should not depend on low-level modules. Both should depend on abstractions.
    2.  Abstractions should not depend on details. Details should depend on abstractions.

This sounds academic, but it's the core of what makes frameworks like NestJS or Angular work. It's about inverting the traditional flow of control.

**Example of a Violation (Traditional Flow):**

High-level policy depends directly on low-level mechanism.

```typescript
// Low-level mechanism
class MySqlDatabase {
  query(sql: string) {
    // ... logic to connect to MySQL and run a query
  }
}

// High-level policy
class UserReportGenerator {
  private db: MySqlDatabase; // <-- Direct dependency on a concrete class

  constructor() {
    this.db = new MySqlDatabase(); // The high-level class controls the dependency
  }

  generateReport() {
    this.db.query('SELECT * FROM users');
    // ... format the report
  }
}
```

**Problems:**
*   The `UserReportGenerator` is **tied to MySQL**. You can't reuse it with Postgres or MongoDB.
*   You can't **test** the `UserReportGenerator` without a running MySQL database. This is a nightmare.

**A Better Approach (DIP Compliant):**

Invert the control! The high-level module defines an interface it needs, and the low-level module implements it.

```typescript
// 1. The abstraction (owned by the high-level layer)
interface Database {
  query(sql: string): any[];
}

// 2. The high-level module depends ONLY on the abstraction
class UserReportGenerator {
  private db: Database;

  // The dependency is "injected" from the outside.
  // This is Dependency Injection!
  constructor(db: Database) {
    this.db = db;
  }

  generateReport() {
    const users = this.db.query('SELECT * FROM users');
    // ... format the report
  }
}

// 3. The low-level details depend on the abstraction
class MySqlDatabase implements Database {
  query(sql: string): any[] {
    console.log('Running MySQL query...');
    return [];
  }
}

class PostgresDatabase implements Database {
  query(sql: string): any[] {
    console.log('Running Postgres query...');
    return [];
  }
}

// For testing, you can create a mock detail
class MockDatabase implements Database {
  query(sql: string): any[] {
    console.log('Running mock query...');
    return [{ id: 1, name: 'Test User' }];
  }
}

// The "Main" or "Composition Root" of your app wires it all up
const mySqlDb = new MySqlDatabase();
const reportGenerator = new UserReportGenerator(mySqlDb);

// For tests:
const mockDb = new MockDatabase();
const testReportGenerator = new UserReportGenerator(mockDb);
```

The direction of dependency has been **inverted**. Instead of `High-level -> Low-level`, it's now `High-level -> Abstraction <- Low-level`. This is the secret sauce for creating decoupled, testable, and maintainable systems.
