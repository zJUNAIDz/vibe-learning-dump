
# Interview Mastery: Glossary of Terms

This glossary provides a quick reference for the key concepts and principles related to Object-Oriented Design and Design Patterns.

---

## A

*   **Abstract Class:** A class that cannot be instantiated on its own and is meant to be subclassed. It can contain both abstract methods (without an implementation) and concrete methods (with an implementation).
*   **Abstract Factory:** A creational pattern that provides an interface for creating families of related or dependent objects without specifying their concrete classes.
*   **Abstraction:** The principle of hiding the complex implementation details and showing only the essential features of the object.
*   **Adapter:** A structural pattern that allows objects with incompatible interfaces to collaborate.

## B

*   **Behavioral Patterns:** Design patterns concerned with algorithms and the assignment of responsibilities between objects. (e.g., Strategy, Observer, Command).
*   **Bridge:** A structural pattern that decouples an abstraction from its implementation so that the two can vary independently.
*   **Builder:** A creational pattern that lets you construct complex objects step by step. The pattern allows you to produce different types and representations of an object using the same construction code.

## C

*   **Caretaker:** A class that is part of the Memento pattern. It is responsible for holding onto a Memento but never inspects or operates on it.
*   **Chain of Responsibility:** A behavioral pattern that lets you pass requests along a chain of handlers. Upon receiving a request, each handler decides either to process the request or to pass it to the next handler in the chain.
*   **Client:** An object that uses another object or pattern.
*   **Cohesion:** A measure of how strongly related and focused the responsibilities of a single module or class are. High cohesion is good.
*   **Command:** A behavioral pattern that turns a request into a stand-alone object that contains all information about the request.
*   **Component:** A class that is part of the Composite pattern. It is the base interface for both leaf and composite objects.
*   **Composite:** A structural pattern that lets you compose objects into tree structures and then work with these structures as if they were individual objects.
*   **Composition:** A "has-a" relationship where one object contains another. The contained object's lifecycle is often managed by the container. (See also: Aggregation).
*   **Composition over Inheritance:** The principle that classes should achieve polymorphic behavior and code reuse by their composition (by containing instances of other classes) rather than inheritance from a base or parent class.
*   **Concrete Class:** A class that can be instantiated and is not abstract.
*   **Context:** A class that uses a Strategy, State, or other pattern. It holds a reference to a strategy or state object and delegates work to it.
*   **Coupling:** The degree of interdependence between software modules. Low coupling is good.
*   **Creational Patterns:** Design patterns that deal with object creation mechanisms, trying to create objects in a manner suitable to the situation. (e.g., Factory, Singleton, Builder).

## D

*   **Decorator:** A structural pattern that lets you attach new behaviors to objects by placing these objects inside special wrapper objects that contain the behaviors.
*   **Dependency Inversion Principle (DIP):** The "D" in SOLID. High-level modules should not depend on low-level modules. Both should depend on abstractions. Abstractions should not depend on details. Details should depend on abstractions.
*   **Dependency Injection (DI):** A technique in which an object receives other objects that it depends on, rather than creating them itself. A specific implementation of Inversion of Control.
*   **Double Dispatch:** A mechanism that dispatches a function call to different concrete functions depending on the runtime types of two objects involved in the call. The core mechanism of the Visitor pattern.

## E

*   **Encapsulation:** The bundling of data with the methods that operate on that data, and the restricting of direct access to some of an object's components.

## F

*   **Facade:** A structural pattern that provides a simplified interface to a library, a framework, or any other complex set of classes.
*   **Factory Method:** A creational pattern that provides an interface for creating objects in a superclass, but lets subclasses alter the type of objects that will be created.
*   **Flyweight:** A structural pattern that lets you fit more objects into the available amount of RAM by sharing common parts of state between multiple objects instead of keeping all of the data in each object.

## I

*   **Inheritance:** An "is-a" relationship where a new class (subclass) is derived from an existing class (superclass), inheriting its fields and methods.
*   **Interface:** A contract that defines a set of methods that a class must implement. It contains no implementation.
*   **Interface Segregation Principle (ISP):** The "I" in SOLID. No client should be forced to depend on methods it does not use. It's better to have many small, specific interfaces than one large, general-purpose interface.
*   **Inversion of Control (IoC):** A broad principle in which the control of object creation and dependency resolution is passed to a container or framework.
*   **Iterator:** A behavioral pattern that lets you traverse elements of a collection without exposing its underlying representation.

## L

*   **Liskov Substitution Principle (LSP):** The "L" in SOLID. Objects of a superclass shall be replaceable with objects of its subclasses without breaking the application.

## M

*   **Mediator:** A behavioral pattern that lets you reduce chaotic dependencies between objects. The pattern restricts direct communications between the objects and forces them to collaborate only via a mediator object.
*   **Memento:** A behavioral pattern that lets you save and restore the previous state of an object without revealing the details of its implementation.

## O

*   **Object-Oriented Programming (OOP):** A programming paradigm based on the concept of "objects", which can contain data and code: data in the form of fields (often known as attributes or properties), and code, in the form of procedures (often known as methods).
*   **Observer:** A behavioral pattern that lets you define a subscription mechanism to notify multiple objects about any events that happen to the object they’re observing.
*   **Open/Closed Principle (OCP):** The "O" in SOLID. Software entities (classes, modules, functions, etc.) should be open for extension, but closed for modification.
*   **Originator:** A class that is part of the Memento pattern. It is the object whose state needs to be saved.

## P

*   **Polymorphism:** The ability of an object to take on many forms. In OOP, it's the ability of a method or property to behave differently based on the object that it is acting upon.
*   **Prototype:** A creational pattern that lets you copy existing objects without making your code dependent on their classes.
*   **Proxy:** A structural pattern that provides a surrogate or placeholder for another object to control access to it.

## R

*   **Receiver:** A class that is part of the Command pattern. It knows how to perform the operations associated with carrying out a request.

## S

*   **Single Responsibility Principle (SRP):** The "S" in SOLID. A class should have only one reason to change, meaning it should have only one job or responsibility.
*   **Singleton:** A creational pattern that ensures a class has only one instance, while providing a global access point to this instance.
*   **SOLID:** An acronym for five design principles intended to make software designs more understandable, flexible, and maintainable: Single Responsibility, Open/Closed, Liskov Substitution, Interface Segregation, and Dependency Inversion.
*   **State:** A behavioral pattern that lets an object alter its behavior when its internal state changes. It appears as if the object changed its class.
*   **Strategy:** A behavioral pattern that lets you define a family of algorithms, put each of them into a separate class, and make their objects interchangeable.
*   **Structural Patterns:** Design patterns that explain how to assemble objects and classes into larger structures, while keeping these structures flexible and efficient. (e.g., Adapter, Composite, Proxy).
*   **Subject:** A class that is part of the Observer pattern. It is the object being observed. It maintains a list of observers and notifies them of state changes.

## T

*   **Template Method:** A behavioral pattern that defines the skeleton of an algorithm in the superclass but lets subclasses override specific steps of the algorithm without changing its structure.

## V

*   **Visitor:** A behavioral pattern that lets you separate algorithms from the objects on which they operate. It allows adding new operations to an object structure without modifying the objects themselves.
