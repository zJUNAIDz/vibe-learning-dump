# Level 06: Type-Level Programming Patterns

## The Problem

TypeScript's structural type system lets bugs through that *feel* wrong:

```typescript
type UserId = string;
type OrderId = string;

function getUser(id: UserId) { /* ... */ }

const orderId: OrderId = 'order_123';
getUser(orderId);  // ✅ No error. UserId and OrderId are both just string.
```

You deleted a user using an order ID. TypeScript didn't catch it because `string === string` structurally.

This level teaches patterns that make **invalid states unrepresentable** at the type level.

---

## Branded / Opaque Types

### The Pattern

Add a phantom property that exists only in the type system:

```typescript
declare const __brand: unique symbol;

type Brand<T, B extends string> = T & { readonly [__brand]: B };

type UserId = Brand<string, 'UserId'>;
type OrderId = Brand<string, 'OrderId'>;

// Constructor functions (the only way to create branded values)
function userId(id: string): UserId {
  // Runtime validation can go here
  return id as UserId;
}

function orderId(id: string): OrderId {
  return id as OrderId;
}

// Now this is a compile error:
function getUser(id: UserId) { /* ... */ }

const oid = orderId('order_123');
getUser(oid);  // ❌ Type 'OrderId' is not assignable to type 'UserId'
```

### Why `unique symbol`?

```typescript
// Without unique symbol, two brands with different names could collide
// if someone creates their own brand with the same property name.
// `unique symbol` guarantees the __brand property is truly unique.
```

### Real-world branded types

```typescript
type Email = Brand<string, 'Email'>;
type URL = Brand<string, 'URL'>;
type PositiveInt = Brand<number, 'PositiveInt'>;
type NonEmptyString = Brand<string, 'NonEmptyString'>;
type Sanitized = Brand<string, 'Sanitized'>;
type ISO8601 = Brand<string, 'ISO8601'>;

// Smart constructors with validation
function email(raw: string): Email {
  if (!raw.includes('@')) throw new Error(`Invalid email: ${raw}`);
  return raw as Email;
}

function positiveInt(n: number): PositiveInt {
  if (!Number.isInteger(n) || n <= 0) throw new Error(`Not a positive integer: ${n}`);
  return n as PositiveInt;
}
```

### Branded types compose

```typescript
type ValidatedEmail = Brand<string, 'Email'> & Brand<string, 'Validated'>;

function validateEmail(input: string): Email {
  if (!input.includes('@')) throw new Error('Invalid');
  return input as Email;
}

function verifyEmail(email: Email): ValidatedEmail {
  // Send verification email, user clicks link
  return email as ValidatedEmail;
}

function sendOnlyToVerified(to: ValidatedEmail, body: string) { /* ... */ }

const raw = 'user@test.com';
const parsed = validateEmail(raw);
sendOnlyToVerified(parsed, 'Hello');  // ❌ Not verified yet

const verified = verifyEmail(parsed);
sendOnlyToVerified(verified, 'Hello');  // ✅
```

**At scale:** Branded types prevent entire categories of bugs — wrong ID types, unsanitized input reaching databases, unvalidated data crossing trust boundaries. The "tax" is constructor functions, which also serve as natural validation points.

---

## Exhaustiveness Checking

### The Pattern

```typescript
function assertNever(x: never): never {
  throw new Error(`Unexpected value: ${JSON.stringify(x)}`);
}
```

If the code reaches `assertNever`, TypeScript will error unless `x` has been narrowed to `never` — meaning all possibilities have been handled.

### Discriminated union + exhaustiveness

```typescript
type Shape =
  | { kind: 'circle'; radius: number }
  | { kind: 'rectangle'; width: number; height: number }
  | { kind: 'triangle'; base: number; height: number };

function area(shape: Shape): number {
  switch (shape.kind) {
    case 'circle':
      return Math.PI * shape.radius ** 2;
    case 'rectangle':
      return shape.width * shape.height;
    case 'triangle':
      return 0.5 * shape.base * shape.height;
    default:
      return assertNever(shape);  // ✅ Compiles — all cases handled
  }
}
```

Now add a new variant:

```typescript
type Shape =
  | { kind: 'circle'; radius: number }
  | { kind: 'rectangle'; width: number; height: number }
  | { kind: 'triangle'; base: number; height: number }
  | { kind: 'polygon'; sides: number; sideLength: number };  // NEW

// Now area() produces a compile error at the assertNever line:
// Type '{ kind: "polygon"; sides: number; sideLength: number }' is not assignable to parameter of type 'never'
```

The compiler **forces** you to handle the new case.

### Exhaustiveness with return types (alternative pattern)

```typescript
// Without assertNever — use return type annotation
function area(shape: Shape): number {
  switch (shape.kind) {
    case 'circle': return Math.PI * shape.radius ** 2;
    case 'rectangle': return shape.width * shape.height;
    // Missing 'triangle' and 'polygon' — TS knows the function might return undefined
    // but return type says number → error
  }
}
```

### Exhaustiveness with objects

```typescript
const areaCalculators: Record<Shape['kind'], (shape: any) => number> = {
  circle: (s) => Math.PI * s.radius ** 2,
  rectangle: (s) => s.width * s.height,
  triangle: (s) => 0.5 * s.base * s.height,
  // If 'polygon' is missing → compile error: Property 'polygon' is missing in type...
};
```

---

## Enforcing Invariants at Compile Time

### Pattern: Mutually exclusive properties

```typescript
type XOR<A, B> =
  | (A & { [K in keyof B]?: never })
  | (B & { [K in keyof A]?: never });

type ByID = { id: string };
type ByEmail = { email: string };

function findUser(query: XOR<ByID, ByEmail>) { /* ... */ }

findUser({ id: '123' });            // ✅
findUser({ email: 'a@b.com' });     // ✅
findUser({ id: '123', email: '' }); // ❌ Can't provide both
findUser({});                        // ❌ Must provide one
```

### Pattern: At-least-one-of

```typescript
type AtLeastOne<T, Keys extends keyof T = keyof T> =
  Keys extends keyof T
    ? Omit<T, Keys> & Required<Pick<T, Keys>>
    : never;

type Filter = AtLeastOne<{
  name?: string;
  email?: string;
  role?: string;
}>;

// Must provide at least one of name, email, or role
```

### Pattern: Dependent types (conditional properties)

```typescript
type Notification =
  | { type: 'email'; address: string; subject: string }
  | { type: 'sms'; phoneNumber: string }
  | { type: 'push'; deviceToken: string; badge?: number };

// Each type has its own required fields. You CANNOT create a notification
// with type 'email' but no address — it's a different union member.

function send(notification: Notification) {
  switch (notification.type) {
    case 'email':
      // notification.address is guaranteed to exist
      sendEmail(notification.address, notification.subject);
      break;
    case 'sms':
      // notification.phoneNumber is guaranteed to exist
      sendSMS(notification.phoneNumber);
      break;
  }
}
```

### Pattern: State machines with types

```typescript
type Draft = { status: 'draft'; content: string };
type Review = { status: 'review'; content: string; reviewer: string };
type Published = { status: 'published'; content: string; publishedAt: Date; url: string };

type Article = Draft | Review | Published;

// Transitions are type-safe functions
function submitForReview(article: Draft, reviewer: string): Review {
  return { status: 'review', content: article.content, reviewer };
}

function publish(article: Review): Published {
  return {
    status: 'published',
    content: article.content,
    publishedAt: new Date(),
    url: `/articles/${encodeURIComponent(article.content.slice(0, 20))}`,
  };
}

// You CANNOT publish a draft — the types prevent it
// publish(draftArticle);  // ❌ Type 'Draft' is not assignable to parameter of type 'Review'
```

---

## Nominal Typing Hacks

Beyond branded types, there are other ways to enforce nominal-like typing:

### Class-based nominal types

```typescript
class Email {
  private readonly __nominal = 'Email';
  constructor(public readonly value: string) {
    if (!value.includes('@')) throw new Error('Invalid email');
  }
}

class PhoneNumber {
  private readonly __nominal = 'PhoneNumber';
  constructor(public readonly value: string) {}
}

function sendEmail(to: Email) { /* ... */ }

sendEmail(new Email('user@test.com'));    // ✅
sendEmail(new PhoneNumber('555-1234'));   // ❌ Different class
```

The `private` field makes these structurally incompatible.

### The `satisfies` + brand combo

```typescript
// Define your schema
const UserSchema = {
  id: 'string' as const,
  name: 'string' as const,
  age: 'number' as const,
} satisfies Record<string, 'string' | 'number' | 'boolean'>;

// Derive the type from the schema
type User = {
  [K in keyof typeof UserSchema]: typeof UserSchema[K] extends 'string' ? string
    : typeof UserSchema[K] extends 'number' ? number
    : typeof UserSchema[K] extends 'boolean' ? boolean
    : never;
};
// { id: string; name: string; age: number }
```

---

## The `never` Type as a Tool

### `never` in mapped types — filtering

```typescript
type MethodKeys<T> = {
  [K in keyof T]: T[K] extends (...args: any[]) => any ? K : never;
}[keyof T];

type Methods<T> = Pick<T, MethodKeys<T>>;

class UserService {
  name = 'UserService';
  getUser(id: string) { return { id }; }
  deleteUser(id: string) { return true; }
}

type ServiceMethods = Methods<UserService>;
// { getUser: (id: string) => { id: string }; deleteUser: (id: string) => boolean }
```

### `never` in unions — identity element

```typescript
type T = string | never;  // string
// never disappears from unions — it's the empty set
```

### `never` in intersections — absorbing element

```typescript
type T = string & never;  // never
// never absorbs in intersections
```

### `never` for impossible states

```typescript
type Direction = 'up' | 'down' | 'left' | 'right';

type Opposite<D extends Direction> =
  D extends 'up' ? 'down' :
  D extends 'down' ? 'up' :
  D extends 'left' ? 'right' :
  D extends 'right' ? 'left' :
  never;  // Exhaustive — if all cases covered, this is unreachable

type T = Opposite<'up'>;  // 'down'
```

---

## Exercises

1. **Build a type-safe FSM (finite state machine):**

Define states and transitions for an order: `pending → paid → shipped → delivered`. Each transition function should accept only the correct input state and return the next state.

2. **Build `ExactlyOne<T>`:**

```typescript
// Given an object type, create a type where exactly one property must be set
// and all others must be absent (not just undefined)
type Lookup = ExactlyOne<{ byId: string; byEmail: string; byName: string }>;

// Valid:
const a: Lookup = { byId: '123' };
const b: Lookup = { byEmail: 'test@test.com' };

// Invalid:
const c: Lookup = { byId: '123', byEmail: 'test@test.com' };  // ❌
const d: Lookup = {};  // ❌
```

3. **Make branded types for a money system:**

Create `USD`, `EUR`, `GBP` branded number types. Write `add(a, b)` that only works if both amounts are the same currency. Write `convert(amount, from, to, rate)` that returns the target currency type.

---

## Next

→ [Level 07: Runtime ↔ Type System Bridge](../07-runtime-type-bridge/)
