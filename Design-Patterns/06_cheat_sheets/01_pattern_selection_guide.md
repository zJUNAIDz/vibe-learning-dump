
# Cheat Sheets: Pattern Selection Guide

Choosing the right design pattern can be tricky. This guide provides a problem-based approach to help you select the most appropriate pattern for your situation.

---

## 1. Creational Patterns: Problems of Object Creation

| If your problem is... | ...consider using the... | Because... |
| :--- | :--- | :--- |
| "I need to create complex objects step-by-step, and the final representation can vary." | **Builder** | It separates the complex construction of an object from its final representation, allowing you to create different variations using the same construction process. |
| "I need to create objects, but I want subclasses to decide which exact class to instantiate." | **Factory Method** | It defines an interface for creating an object, but lets subclasses alter the type of objects that will be created. It's about deferring instantiation to subclasses. |
| "I need to create families of related objects without specifying their concrete classes." | **Abstract Factory** | It provides an interface for creating families of related objects (e.g., a `MacOSFactory` that creates `MacOSButton` and `MacOSCheckbox`). |
| "I need to copy an existing object, but I don't want my code to depend on its class." | **Prototype** | It allows you to create new objects by copying a "prototype" instance, which is often faster and more convenient than construction from scratch. |
| "I need to ensure that a class has only one instance and provide a global point of access to it." | **Singleton** | It guarantees a single instance and provides a global access point, useful for things like loggers, database connections, or hardware interface access. Use with caution. |

---

## 2. Structural Patterns: Problems of Object Composition

| If your problem is... | ...consider using the... | Because... |
| :--- | :--- | :--- |
| "I need to make two objects with incompatible interfaces work together." | **Adapter** | It acts as a wrapper or translator between two incompatible interfaces. |
| "I need to separate an abstraction from its implementation so they can evolve independently." | **Bridge** | It splits a large class or a set of closely related classes into two separate hierarchies—abstraction and implementation—which can be developed independently of each other. |
| "I need to treat a group of objects in the same way as a single object." | **Composite** | It lets you compose objects into tree structures and then work with these structures as if they were individual objects. (e.g., a file system with files and directories). |
| "I need to add new responsibilities to an object dynamically without subclassing." | **Decorator** | It lets you attach new behaviors to objects by placing them inside special wrapper objects that contain the behaviors. |
| "I need to provide a simple, unified interface to a complex subsystem." | **Facade** | It provides a simplified interface to a library, a framework, or any other complex set of classes, hiding its complexity. |
| "I have a huge number of objects that are draining memory, and many of them have duplicate state." | **Flyweight** | It minimizes memory usage by sharing as much data as possible with other similar objects; it's a way to manage the state of many objects efficiently. |
| "I need to control access to an object, or manage its creation/loading." | **Proxy** | It provides a surrogate or placeholder for another object to control access to it. Used for lazy initialization (Virtual Proxy), access control (Protection Proxy), or logging (Logging Proxy). |

---

## 3. Behavioral Patterns: Problems of Object Communication

| If your problem is... | ...consider using the... | Because... |
| :--- | :--- | :--- |
| "I have a request that could be handled by several different objects, and I don't want the sender to be coupled to the receivers." | **Chain of Responsibility** | It passes a request along a chain of handlers. Each handler decides either to process the request or to pass it to the next handler. (e.g., middleware). |
| "I need to issue requests to objects without knowing anything about the operation being requested or the receiver of the request." | **Command** | It turns a request into a stand-alone object. This lets you parameterize clients with different requests, queue requests, and support undoable operations. |
| "I need to traverse a collection of objects without exposing its internal structure." | **Iterator** | It provides a uniform way to access the elements of a collection sequentially without exposing its underlying representation. |
| "I have a set of objects that communicate in a complex, many-to-many way, and it's becoming a tangled mess." | **Mediator** | It reduces chaotic dependencies by forcing objects to communicate only through a central mediator object. |
| "I need to save and restore the state of an object without violating its encapsulation." | **Memento** | It lets you capture and externalize an object's internal state so that the object can be restored to this state later. (e.g., undo/redo). |
| "I have an object that needs to notify other objects when its state changes, without being coupled to them." | **Observer** | It defines a subscription mechanism to notify multiple objects about any events that happen to the object they’re observing. (e.g., event listeners). |
| "An object's behavior depends on its state, and it must change its behavior at runtime depending on that state." | **State** | It allows an object to alter its behavior when its internal state changes. It's an object-oriented way to implement a state machine. |
| "I need to let an object choose its behavior from a family of algorithms at runtime." | **Strategy** | It defines a family of algorithms, puts each of them into a separate class, and makes their objects interchangeable. |
| "I have an algorithm with a fixed structure, but I want subclasses to be able to change specific steps." | **Template Method** | It defines the skeleton of an algorithm in a superclass but lets subclasses override specific steps of the algorithm without changing its structure. (Based on inheritance). |
| "I need to add new operations to a set of classes without changing the classes themselves." | **Visitor** | It lets you separate an algorithm from an object structure on which it operates. It's great for stable object hierarchies where you need to frequently add new operations. |
