
# Cheat Sheets: SOLID Principles Summary

SOLID is a mnemonic acronym for five design principles that are fundamental to good object-oriented design. Adhering to these principles helps create systems that are more understandable, flexible, and maintainable.

---

## S: Single Responsibility Principle (SRP)

*   **The Gist:** A class should have only one reason to change.
*   **In Simple Terms:** Every class should have one, and only one, job.
*   **Why?**
    *   **Maintainability:** When a class has only one responsibility, it's smaller and easier to understand. Changes are isolated. If you need to change how users are authenticated, you only have to look in the `UserAuthenticator` class, not in a massive `User` class that also knows how to save to a database and render to HTML.
    *   **Testability:** A class with one job is much easier to test.
*   **Smell / Violation:** A "God Object"—a class that knows and does everything. Methods that seem out of place. A class that needs to be changed for many different reasons (e.g., changing a business rule, changing a database schema, and changing a UI presentation all require changes to the same class).
*   **Example:** Instead of a `Report` class that both calculates the data and formats it as JSON, you should have a `ReportCalculator` class and a `ReportJsonFormatter` class.

---

## O: Open/Closed Principle (OCP)

*   **The Gist:** Software entities (classes, modules, functions) should be open for extension, but closed for modification.
*   **In Simple Terms:** You should be able to add new functionality without changing existing code.
*   **Why?**
    *   **Stability:** Changing existing, working code is risky. It can introduce bugs into features that were previously stable.
    *   **Flexibility:** It allows you to extend a system's behavior without having to re-test everything.
*   **How?** Through abstraction. Use interfaces, abstract classes, and patterns like **Strategy**, **Template Method**, or **Decorator**.
*   **Smell / Violation:** A massive `if/else` or `switch` statement that you have to modify every time a new "type" is added. For example, a `calculateArea(shape)` function that has a `switch` on the shape's type. To add a new shape, you have to modify this function.
*   **Example:** Instead of the `switch` statement, make `shape` an interface with a `getArea()` method. Now, you can add new shapes (new classes implementing the interface) without ever touching the original code.

---

## L: Liskov Substitution Principle (LSP)

*   **The Gist:** Subtypes must be substitutable for their base types.
*   **In Simple Terms:** If you have a function that works with a base class `Animal`, it should also work with any of `Animal`'s subclasses (like `Dog` or `Cat`) without any issues.
*   **Why?**
    *   **Reliability:** It ensures that your inheritance hierarchies are correct and that polymorphism works as expected. Violating LSP leads to unexpected behavior and runtime errors.
*   **Smell / Violation:** A subclass that overrides a base class method and does something completely different, throws an unexpected exception, or has a more restrictive validation rule. The classic (though debated) example is a `Rectangle` class with a `Square` subclass. If you have `set_width(w)` and `set_height(h)` methods, a `Square` must keep width and height equal, which a `Rectangle` does not. This violates the "behavior" of the base class.
*   **Example:** A `Bird` class has a `fly()` method. A `Penguin` subclass might override `fly()` to throw an "I can't fly!" exception. This violates LSP because code that works with a `Bird` would break if you gave it a `Penguin`. The correct design is to have a more specific `FlyingBird` interface/subclass.

---

## I: Interface Segregation Principle (ISP)

*   **The Gist:** No client should be forced to depend on methods it does not use.
*   **In Simple Terms:** Keep your interfaces small and focused. It's better to have many small interfaces than one big one.
*   **Why?**
    *   **Decoupling:** It prevents "fat" interfaces that lead to unnecessary dependencies. If a class implements an interface with methods it doesn't need, it's forced to implement them (often as empty methods), and it's coupled to changes in methods it doesn't even use.
*   **Smell / Violation:** A class that implements an interface but leaves some of its methods empty or throwing "not implemented" exceptions.
*   **Example:** A single `IWorker` interface has methods `work()` and `eat()`. A `RobotWorker` class can implement `work()` but has no use for `eat()`. The solution is to segregate the interface into `IWorkable` and `IEatable`. A `HumanWorker` can implement both, while a `RobotWorker` only needs to implement `IWorkable`.

---

## D: Dependency Inversion Principle (DIP)

*   **The Gist:**
    1.  High-level modules should not depend on low-level modules. Both should depend on abstractions.
    2.  Abstractions should not depend on details. Details should depend on abstractions.
*   **In Simple Terms:** Don't let your high-level business logic depend on low-level implementation details. Instead, make both depend on interfaces that you control.
*   **Why?**
    *   **Decoupling & Testability:** This is the key principle that enables Dependency Injection (DI). When your `UserService` (high-level) depends on an `IUserRepository` interface (abstraction) instead of a concrete `PostgresUserRepository` (low-level), you have "inverted" the dependency. This allows you to easily swap out the `PostgresUserRepository` for a `MongoUserRepository` or a `MockUserRepository` for testing without changing the `UserService` at all.
*   **Smell / Violation:** Using the `new` keyword to create a dependency inside a class. For example, `this.userRepository = new PostgresUserRepository();` inside a service class.
*   **Example:**
    *   **Violation:** `class UserService { constructor() { this.repo = new PostgresRepo(); } }`
    *   **Adherence:** `class UserService { constructor(repo: IRepo) { this.repo = repo; } }`
