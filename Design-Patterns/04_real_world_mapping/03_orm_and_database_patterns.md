
# Real-World Mapping: ORM and Database Patterns

Object-Relational Mapping (ORM) libraries like Prisma, TypeORM, Hibernate (Java), and Entity Framework (.NET) are essential tools in modern application development. They solve the "object-relational impedance mismatch"—the problem that the way we think about objects in our code (with inheritance, collections, and complex relationships) is very different from the way data is stored in a relational database (in flat, normalized tables).

ORMs are themselves complex pieces of software built on a foundation of design patterns. Understanding these patterns helps demystify how ORMs work and how to use them effectively.

---

## 1. 🗄️ The Repository and Unit of Work Patterns

These two patterns are the heart of most ORMs' interaction models.

*   **The Problem:** You don't want your business logic (e.g., a `UserService`) to be littered with raw SQL queries. This couples your logic directly to the database schema and the specific SQL dialect, making it hard to maintain and test.

*   **The Solution (Repository Pattern):**
    *   A **Repository** is an object that encapsulates the logic for retrieving, storing, and deleting a specific type of entity (like a `User` or a `Product`). It provides a simple, collection-like interface.
    *   It acts as a **Facade** over the underlying data store. Your service layer doesn't know if the data is coming from PostgreSQL, MongoDB, or an in-memory array for testing.

    ```typescript
    // Your service only knows about the repository interface.
    class UserService {
      constructor(private readonly userRepository: IUserRepository) {}

      async create(name: string) {
        const user = new User(name);
        // The service uses a simple, domain-focused interface.
        await this.userRepository.add(user);
      }
    }

    // The repository implementation has the data-access logic.
    class PrismaUserRepository implements IUserRepository {
      async add(user: User) {
        // It translates domain objects into database operations.
        await prisma.user.create({ data: user });
      }
    }
    ```

*   **The Solution (Unit of Work Pattern):**
    *   A **Unit of Work** keeps track of everything you do during a business transaction that can affect the database. When you're done, it figures out everything that needs to be changed and writes it all out in a single transaction.
    *   This ensures data consistency and minimizes database round-trips.

    ```typescript
    // The Unit of Work tracks changes.
    async function changeUsername(userId: number, newName: string) {
      const uow = new UnitOfWork();
      const userRepo = uow.getUserRepository();
      const auditRepo = uow.getAuditRepository();

      const user = await userRepo.findById(userId);
      user.name = newName; // The UoW tracks this change in memory.

      const auditLog = new AuditLog(`Changed name for user ${userId}`);
      auditRepo.add(auditLog); // The UoW tracks this new object.

      // All changes are saved in a single transaction.
      await uow.commit();
    }
    ```
    Most modern ORMs (like Entity Framework and Hibernate) have a built-in Unit of Work. Their `DbContext` or `Session` object automatically tracks all changes made to entities and saves them in a single transaction when you call `saveChanges()` or `flush()`.

---

## 2. 🕵️ The Identity Map Pattern

*   **The Problem:** In a single business transaction, you might request the same entity multiple times. For example, fetching user `123` and then fetching the author of a post written by user `123`. If the ORM creates a new `User` object every time, you have two different objects in memory representing the same database row. This wastes memory and can lead to data inconsistency if you modify one but not the other.

*   **The Solution (Identity Map):**
    *   The ORM maintains a session-level cache (a map) of all entities that have been retrieved from the database. The map's keys are the entity type and primary key (e.g., `['User', 123]`).
    *   When you request an entity, the ORM first checks the Identity Map.
    *   If the entity is in the map, it returns the existing object instance.
    *   If not, it fetches the data from the database, creates a new entity object, stores it in the map, and then returns it.
    *   This is an application of the **Flyweight** or **Singleton** (scoped to a session) pattern. It guarantees that for a given session, there is only one object instance for each database row.

---

## 3. 👻 The Virtual Proxy Pattern (Lazy Loading)

*   **The Problem:** You fetch a `User` object. That user has 5,000 `Order` records associated with it. If the ORM loads all 5,000 orders every time you fetch a user, your application will be incredibly slow and memory-intensive. This is the "N+1 query problem" in another form.

*   **The Solution (Lazy Loading via Virtual Proxy):**
    *   When the ORM creates the `User` object, it doesn't fetch the related orders. Instead, for the `user.orders` property, it assigns a special, dynamically-generated **Proxy** object.
    *   This proxy object looks and feels exactly like a collection of `Order`s, but it's empty.
    *   The first time your code tries to access the `user.orders` property (e.g., by calling `user.orders.length` or iterating over it), the proxy "wakes up."
    *   The proxy, which has a reference to the ORM's session, now executes the database query to fetch the actual orders.
    *   It replaces itself with the real collection of orders and returns them to your code.
    *   Subsequent accesses to `user.orders` hit the now-loaded collection directly.
    *   This is a perfect implementation of the **Proxy** pattern, specifically a Virtual Proxy. It defers the expensive operation until it's absolutely necessary.

---

## Summary: How ORMs Use Patterns

| Pattern | How it's used in an ORM |
| :--- | :--- |
| **Repository** | Provides a collection-like interface for data access, abstracting away SQL. |
| **Unit of Work** | Tracks changes to entities and commits them all in a single database transaction. |
| **Facade** | The Repository and Unit of Work together act as a Facade over the complexity of database interaction. |
| **Identity Map** | A session-level cache that ensures only one object instance exists per database row, per session. (A type of Flyweight). |
| **Proxy** | Used to implement "lazy loading" for related entities, deferring expensive queries until the data is actually accessed. |
| **Adapter** | The ORM's database drivers act as Adapters, translating the ORM's standard query language into the specific SQL dialect of different databases (PostgreSQL, MySQL, etc.). |
| **Memento** | The Unit of Work can be seen as using the Memento pattern. Before making changes, it knows the original state of the entities. This allows it to generate the correct `UPDATE` statements or even roll back changes in memory. |
