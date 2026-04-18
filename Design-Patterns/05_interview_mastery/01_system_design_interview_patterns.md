
# Interview Mastery: System Design Patterns

In a system design interview, you're asked to design a large-scale service like Twitter, Netflix, or Dropbox. The interviewer isn't looking for a perfect, detailed architecture. They are looking to see how you think about trade-offs, scalability, and reliability.

Design patterns are your best friends in this scenario. They are proven solutions to common problems. Being able to name a pattern, explain its trade-offs, and apply it to the problem at hand shows that you have a deep understanding of software architecture.

Here are some of the most important architectural patterns and how they relate to the classic GoF design patterns.

---

## 1. 📢 Publish-Subscribe (Pub/Sub) Pattern

*   **What it is:** A messaging pattern where senders of messages (Publishers) do not program the messages to be sent directly to specific receivers (Subscribers). Instead, publishers categorize messages into topics, and subscribers subscribe to the topics they are interested in. A central message broker (like RabbitMQ, Kafka, or Google Pub/Sub) is responsible for routing messages from publishers to subscribers.

*   **When to use it:**
    *   **Decoupling services:** When you want to allow services to communicate without knowing about each other. For example, in an e-commerce app, when an order is placed, the `OrderService` can publish an `OrderPlaced` event. The `NotificationService`, `InventoryService`, and `ShippingService` can all subscribe to this event and react accordingly, without the `OrderService` needing to know they exist.
    *   **Asynchronous workflows:** For tasks that can be done in the background, like sending an email, transcoding a video, or generating a report.

*   **GoF Pattern Connection:** This is a large-scale implementation of the **Observer Pattern**. The message broker is the `Subject`, and the services are the `Observers`.

---

## 2. 🗄️ CQRS (Command Query Responsibility Segregation)

*   **What it is:** A pattern that separates the models used for updating data (Commands) from the models used for reading data (Queries). In many systems, the way you need to write data is very different from the way you need to read it. For example, reads might need to be highly optimized with complex joins and denormalized data, while writes need to be normalized and consistent.

*   **How it works:**
    *   **Command Side:** Handles `create`, `update`, and `delete` requests. This side is optimized for writing and consistency. It might use a normalized SQL database.
    *   **Query Side:** Handles `read` requests. This side is optimized for reading. It might use a denormalized read model, a search index like Elasticsearch, or a document database.
    *   The two sides are kept in sync, usually through events (using a Pub/Sub system). When the command side writes data, it publishes an event, and the query side updates its read model.

*   **When to use it:**
    *   In systems with a high number of reads compared to writes.
    *   When you have complex queries that are slow to run on a normalized, write-optimized database.
    *   In collaborative domains where multiple actors are working on the same data.

*   **GoF Pattern Connection:**
    *   **Command Pattern:** The "Command" in CQRS is a direct application of the GoF Command pattern. It encapsulates a request to change state as an object.
    *   **Observer Pattern:** Used to synchronize the write model with the read model.

---

## 3. 🚪 API Gateway

*   **What it is:** A single entry point for all clients. Instead of clients calling dozens of different microservices directly, they make one call to the API Gateway. The gateway then routes the request to the appropriate downstream service(s), aggregates the results, and returns them to the client.

*   **When to use it:**
    *   In a microservices architecture. It simplifies the client by providing a single endpoint and hides the complexity of the backend service topology.
    *   To handle cross-cutting concerns like authentication, rate limiting, logging, and caching in a centralized place.

*   **GoF Pattern Connection:**
    *   **Facade Pattern:** The API Gateway is a classic Facade. It provides a simplified, unified interface to a more complex subsystem (the fleet of microservices).
    *   **Proxy Pattern:** The gateway acts as a reverse proxy, forwarding requests on behalf of the client. It can also be a **Decorator**, adding functionality like authentication or logging to the request before forwarding it.

---

## 4. 📦 Sidecar Pattern

*   **What it is:** A pattern used in containerized environments (like Kubernetes) where you deploy a secondary container (the "sidecar") alongside your main application container. The sidecar's purpose is to augment or enhance the main application by providing supporting features.

*   **When to use it:**
    *   **Service Mesh:** A common use is for networking. A sidecar proxy (like Envoy or Linkerd) can be deployed next to every service to handle things like service discovery, traffic management, retries, and circuit breaking.
    *   **Logging and Monitoring:** A sidecar can collect logs and metrics from the main application and forward them to a centralized logging/monitoring system.

*   **GoF Pattern Connection:**
    *   **Decorator Pattern:** The sidecar "decorates" the main application with additional functionality without being part of the application's own code.
    *   **Proxy Pattern:** A networking sidecar is a proxy that intercepts all incoming and outgoing network traffic for the main container.

---

## 5. 💔 Circuit Breaker Pattern

*   **What it is:** A pattern used to prevent a network or service failure from cascading to other services. When a service calls another service that is failing, the circuit breaker "trips" and stops making further calls to the failing service for a period of time. Instead, it returns an error immediately. After a timeout, it will allow a limited number of "trial" requests. If those succeed, it closes the circuit and resumes normal operation.

*   **When to use it:**
    *   In any distributed system where services make synchronous calls to other services. It improves stability and resilience by preventing a single failing service from bringing down the entire system.

*   **GoF Pattern Connection:**
    *   **State Pattern:** A circuit breaker is a state machine. It can be in one of three states: `Closed` (normal operation), `Open` (calls fail immediately), or `Half-Open` (allowing a trial request). The behavior of the circuit breaker changes based on its current state.
    *   **Proxy Pattern:** The circuit breaker is often implemented as a Proxy that wraps the remote service call. The proxy intercepts the call and decides whether to let it through, fail it immediately, or send a trial request, based on its current state.
