
# 🤥 Common OOP Lies Your Professor Told You

Alright, you've had the Kool-Aid. You've seen the "four pillars" and the neat little `Drone` class. It all looks so clean, so perfect. Now, let's talk about the real world, where theory gets mugged in a dark alley by reality.

Many of the "rules" you learned about OOP are, at best, oversimplifications and, at worst, outright lies that lead to terrible, brittle, and unmaintainable code.

---

## Lie #1: "Model the Real World"

This is the most pervasive and damaging lie in OOP education. The idea that you should create a class for every "noun" in your problem description. `Customer`, `Product`, `Order`. It sounds intuitive, but it's a trap.

**The Problem:** Your software is not the real world. The real world is messy, infinitely complex, and has no single "purpose." Your software, on the other hand, is a **purpose-built system of behaviors**.

When you model the "thing" itself, you end up with anemic objects that are just bags of data. The actual business logic—the *verbs*—gets scattered all over the place in "manager" or "service" classes.

**The Naive "Real World" Model:**

```typescript
class Order {
  // Just a data bucket
  orderId: string;
  items: any[];
  total: number;
  customerId: string;
  status: 'pending' | 'shipped' | 'delivered';
}

class OrderProcessor {
  // All the logic lives here
  shipOrder(order: Order) {
    // ... connect to shipping API
    // ... send email to customer
    // ... update database
    order.status = 'shipped';
  }

  calculateTotal(order: Order) {
    // ... logic to sum up item prices
  }
}
```

This is just procedural programming in disguise. You've separated the data (`Order`) from the logic (`OrderProcessor`), which is the exact problem OOP was supposed to solve! Congrats, you just reinvented a bug factory.

**The Truth:** Don't model nouns. **Model behaviors and responsibilities.**

Instead of an `Order` that just holds data, think about what an `Order` *does*. It can be placed, it can be shipped, it can be cancelled.

**A Better Approach:**

```typescript
class Order {
  private orderId: string;
  private items: any[];
  private status: 'pending' | 'shipped' | 'delivered';

  // The logic lives WITH the data it protects.
  // This is ENCAPSULATION.
  ship() {
    if (this.status !== 'pending') {
      throw new Error("Can't ship an order that isn't pending.");
    }
    // ... internal logic for shipping
    this.status = 'shipped';
    // ... maybe return a TrackingEvent object
  }

  // The status is private. No one can just set it to whatever they want.
  getStatus() {
    return this.status;
  }
}
```

Your objects should be **rich with behavior**, not anemic data structures.

---

## Lie #2: "Inheritance is the Primary Tool for Code Reuse"

This is the siren song of OOP. It looks so elegant. `class SuperDrone extends Drone`. You get all that code for free! What could go wrong?

Everything. Everything can go wrong.

**The Problem:** Inheritance creates the **tightest coupling** possible between two classes. The child class is intimately tied to the parent's implementation details. When the parent changes, the child breaks. This is the "fragile base class" problem.

**Example of Inheritance Hell:**

Imagine a `Bird` class.

```typescript
class Bird {
  fly() {
    console.log('I am flying!');
  }
}

class Duck extends Bird {
  quack() {
    console.log('Quack!');
  }
}
```

Looks great. Now, let's add a `Penguin`. A penguin is a bird, right?

```typescript
class Penguin extends Bird {
  // Uh oh.
  // Penguins don't fly.
  // What do we do?

  // Option 1: Override and do nothing.
  fly() {
    // silent failure... yikes
  }

  // Option 2: Override and throw an error.
  fly() {
    throw new Error('Sorry, I am a flightless bird.');
  }
}

const birds: Bird[] = [new Duck(), new Penguin()];

for (const bird of birds) {
  bird.fly(); // BOOM! Runtime error.
}
```

This is a classic violation of the Liskov Substitution Principle (we'll get to that). You can no longer treat a `Penguin` as a `Bird` everywhere, which defeats the purpose of polymorphism.

**The Truth:** **Favor Composition over Inheritance.**

Instead of saying a `Penguin` *is a* `Bird`, you should say it *has* certain behaviors.

**A Better Approach (Composition):**

```typescript
// Define behaviors as separate, reusable pieces
interface FlyBehavior {
  fly(): void;
}

class CanFly implements FlyBehavior {
  fly() { console.log('I am flying!'); }
}

class CantFly implements FlyBehavior {
  fly() { /* Do nothing or throw error, but it's an explicit choice */ }
}

// The "Bird" is now just a composition of behaviors
class Bird {
  private flyBehavior: FlyBehavior;

  constructor(flyBehavior: FlyBehavior) {
    this.flyBehavior = flyBehavior;
  }

  performFly() {
    this.flyBehavior.fly();
  }
}

// You compose the object you need at runtime
const duck = new Bird(new CanFly());
const penguin = new Bird(new CantFly());

duck.performFly();    // "I am flying!"
penguin.performFly(); // (nothing happens)
```

This is more flexible, less coupled, and won't blow up in your face six months from now.

---

## Lie #3: "Getters and Setters are Good Encapsulation"

Professors love teaching this. Take every private field and expose it with a `get...()` and `set...()` method.

```typescript
class User {
  private name: string;
  private email: string;

  // This is NOT encapsulation. This is a lie.
  public getName(): string {
    return this.name;
  }
  public setName(name: string): void {
    this.name = name;
  }
  public getEmail(): string {
    return this.email;
  }
  public setEmail(email: string): void {
    this.email = email;
  }
}
```

**The Problem:** You haven't encapsulated anything! You've just made it slightly more annoying to access the data. You've exposed the internal structure of your class to the entire world. Any other class can still get the data, change it, and put your `User` object into any state it wants.

**The Truth:** Encapsulation isn't about hiding data; it's about **protecting the integrity of the object**. Expose behaviors, not fields.

Ask yourself *why* something needs to change.

*   Does the user's name need to change? Probably not after creation.
*   Does the user's email need to change? Yes, but is it just a simple assignment? Or does it involve a verification process?

**A Better Approach:**

```typescript
class UserProfile {
  private readonly name: string; // Can't be changed after creation
  private email: string;
  private isEmailVerified: boolean;

  constructor(name: string, email: string) {
    this.name = name;
    this.email = email;
    this.isEmailVerified = false;
  }

  // Expose a BEHAVIOR, not a setter.
  changeEmail(newEmail: string) {
    if (this.email === newEmail) return;

    this.email = newEmail;
    this.isEmailVerified = false; // The business rule!
    // ... maybe trigger a verification email
  }

  // Only expose what's necessary
  getEmail() {
    return this.email;
  }

  isVerified() {
    return this.isEmailVerified;
  }
}
```

See the difference? We're not just setting a field. We're executing a business process (`changeEmail`) that enforces rules and maintains a valid state. That's real encapsulation.
