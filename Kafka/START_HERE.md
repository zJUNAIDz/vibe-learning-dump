# Start Here

## The Problem

You're a backend engineer. You build services. You connect them with HTTP calls and database writes. It works — until it doesn't.

Here's a story:

> A user places an order. Your Order Service saves it to the database, then calls the Payment Service over HTTP. The Payment Service charges the card, then calls the Notification Service to send a confirmation email. The Notification Service calls the Inventory Service to decrement stock.

Everything is synchronous. Everything is coupled. Everything works — on a good day.

Then:
- The Payment Service is slow. The user stares at a spinner for 8 seconds.
- The Notification Service is down. The entire order fails — even though the payment succeeded.
- You deploy a new version of Inventory Service. Orders start failing because the API changed.
- Traffic spikes. Payment Service can't keep up, so orders start timing out.

You've built a distributed monolith. Every service depends on every other service being up, fast, and compatible — right now, at this moment.

## The Question

What if services didn't need to talk to each other directly?

What if the Order Service just wrote down "an order happened" — and other services picked it up whenever they were ready?

What if a service could crash, restart, and pick up exactly where it left off?

What if you could replay the last hour of events to debug a production issue?

That's what Kafka gives you. Not magic. Just an append-only log that decouples services in time.

## How This Curriculum Works

We build one system: an **Order & Payment Pipeline**.

- Phase 0: We build it the naive way — synchronous HTTP calls. We see it break.
- Phase 1: We introduce Kafka as a replacement for direct service calls.
- Phase 2–3: We scale it with partitions and consumer groups.
- Phase 4–5: We handle failures, retries, and poison messages.
- Phase 6: We add schema contracts so services can evolve independently.
- Phase 7: We use Kafka's retention and replay to rebuild state.
- Phase 8: We learn to operate it — monitoring, lag, backpressure.

Every phase extends the same codebase. Every concept is introduced because the project demands it.

## Begin

Go to [Phase 0 — The Pre-Kafka Pain](phase-00-pre-kafka/README.md).
