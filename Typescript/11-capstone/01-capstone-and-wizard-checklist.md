# Level 11: Capstone & TypeScript Wizard Checklist

---

## Capstone Projects

These projects are designed to force deep engagement with every level of this curriculum. Each one has a real-world context and requires combining multiple type-level techniques.

---

### Project 1: Type-Safe HTTP Client

**Build an HTTP client where route types are inferred from a route definition object.**

Requirements:

```typescript
const api = createClient({
  'GET /users': {
    response: z.array(UserSchema),
  },
  'GET /users/:id': {
    params: z.object({ id: z.string() }),
    response: UserSchema,
  },
  'POST /users': {
    body: z.object({ name: z.string(), email: z.string().email() }),
    response: UserSchema,
  },
  'PUT /users/:id': {
    params: z.object({ id: z.string() }),
    body: z.object({ name: z.string().optional() }),
    response: UserSchema,
  },
} as const);

// Usage — fully type-safe, zero annotations:
const users = await api('GET /users');
// type: User[]

const user = await api('GET /users/:id', { params: { id: '123' } });
// type: User
// ❌ api('GET /users/:id')  → error: params required

const created = await api('POST /users', { body: { name: 'Alice', email: 'a@b.com' } });
// type: User
// ❌ api('POST /users', { body: { name: 123 } })  → error: name must be string
```

**Skills tested:**
- Template literal type parsing (extract HTTP method + path)
- Route parameter extraction (`:id` → `{ id: string }`)
- Conditional required/optional arguments based on schema presence
- `z.infer` integration
- Const type parameter inference

---

### Project 2: Type-Safe State Machine

**Build a finite state machine where transitions are enforced at the type level.**

Requirements:

```typescript
const machine = defineMachine({
  initial: 'idle',
  states: {
    idle: { on: { START: 'loading' } },
    loading: { on: { SUCCESS: 'success', FAILURE: 'error' } },
    success: { on: { RESET: 'idle' } },
    error: { on: { RETRY: 'loading', RESET: 'idle' } },
  },
});

let state = machine.start();
// type: { current: 'idle' }

state = machine.send(state, 'START');
// type: { current: 'loading' }

state = machine.send(state, 'SUCCESS');
// type: { current: 'success' }

// ❌ machine.send(state, 'FAILURE')
// Error: event 'FAILURE' is not valid in state 'success'
```

**Skills tested:**
- Discriminated unions
- Mapped types for transition tables
- Generic state tracking through function calls
- `const` type parameter for literal inference
- Exhaustiveness checking

---

### Project 3: Type-Safe SQL Query Builder

**Build a query builder that tracks selected columns, joins, and conditions at the type level.**

Requirements:

```typescript
type DB = {
  users: { id: number; name: string; email: string; teamId: number };
  teams: { id: number; name: string; plan: 'free' | 'pro' };
  posts: { id: number; title: string; authorId: number; published: boolean };
};

const query = db.from('users')
  .select('name', 'email')
  .where('teamId', '=', 1)
  .join('teams', 'users.teamId', 'teams.id')
  .select('teams.plan')
  .orderBy('name');

const results = await query.execute();
// type: { name: string; email: string; plan: 'free' | 'pro' }[]

// ❌ .select('nonexistent')       → error
// ❌ .where('name', '=', 123)     → error: name is string, not number
// ❌ .orderBy('unpickedColumn')   → error
```

**Skills tested:**
- Builder pattern with accumulating generics
- Intersection types for joins
- Key remapping for `table.column` syntax
- Conditional types for operator validation
- Template literal types for column references

---

### Project 4: Zod From Scratch (Mini Version)

**Build a simplified schema validation library with full type inference.**

Requirements:

```typescript
const UserSchema = s.object({
  name: s.string().min(1),
  age: s.number().int().min(0),
  role: s.enum(['admin', 'user'] as const),
  address: s.object({
    street: s.string(),
    zip: s.string().regex(/^\d{5}$/),
  }).optional(),
});

type User = Infer<typeof UserSchema>;
// { name: string; age: number; role: 'admin' | 'user'; address?: { street: string; zip: string } }

const result = UserSchema.safeParse(unknownData);
if (result.ok) {
  result.data;  // User
} else {
  result.errors;  // structured error info
}
```

**Skills tested:**
- Phantom types for type tracking
- Chained method types
- `infer` for type extraction
- Optional/required handling
- Enum literal preservation
- Runtime validation + type derivation

---

### Project 5: Type-Level JSON Parser

**Parse a JSON string at the type level (this is the "final boss").**

```typescript
type Result = ParseJSON<'{"name": "Alice", "age": 30, "active": true}'>;
// type: { name: "Alice"; age: 30; active: true }

type Result2 = ParseJSON<'[1, 2, 3]'>;
// type: [1, 2, 3]

type Result3 = ParseJSON<'"hello"'>;
// type: "hello"
```

**Skills tested:**
- Recursive template literal parsing
- Multiple `infer` positions
- Type-level string scanning
- Conditional types at maximum complexity
- TypeScript recursion depth management

---

## TypeScript Wizard Checklist

You have completed this curriculum when you can confidently check every item:

### Foundation (Levels 00-01)
- [ ] I can explain why TypeScript is two separate languages
- [ ] I can predict when excess property checking applies and when it doesn't
- [ ] I can explain structural vs nominal typing and why TypeScript chose structural
- [ ] I know at least 4 places where TypeScript is intentionally unsound
- [ ] I understand types as sets — union as ∪, intersection as ∩, extends as ⊆
- [ ] I can explain why `{ a: string; b: number }` is a subtype of `{ a: string }`

### Inference (Level 02)
- [ ] I can explain bidirectional (contextual) typing with a real example
- [ ] I know when and why widening occurs
- [ ] I use `as const`, `satisfies`, and `const` type parameters appropriately
- [ ] I can identify and prevent inference loss in my code
- [ ] I almost never annotate types that TypeScript can infer

### Generics (Level 03)
- [ ] I can design generic functions where `T` is inferred from arguments
- [ ] I understand partial inference failure and know the curried function workaround
- [ ] I can use `NoInfer<T>` to control which arguments contribute to inference
- [ ] I can build APIs where callbacks are fully typed from configuration objects
- [ ] I never write unnecessary generics

### Conditional Types (Level 04)
- [ ] I can explain distributive conditional types without looking it up
- [ ] I know how to prevent distribution and when to do it
- [ ] I can use `infer` to extract types from complex structures
- [ ] I understand covariant vs contravariant `infer` positions
- [ ] I can write recursive conditional types with proper base cases

### Mapped Types (Level 05)
- [ ] I know the 3 `keyof` traps (index signatures, unions, intersections)
- [ ] I can use key remapping (`as` clause) to rename, filter, and transform keys
- [ ] I can build deep transformations (DeepPartial, DeepReadonly)
- [ ] I use template literal types for string manipulation at the type level
- [ ] I remember to handle built-in types (Function, Date, Map) in recursive types

### Type-Level Programming (Level 06)
- [ ] I can implement branded types and explain when they're worth the overhead
- [ ] I use exhaustiveness checking (`assertNever`) consistently
- [ ] I can model state machines with discriminated unions
- [ ] I can enforce mutual exclusion, at-least-one, and other invariants at compile time
- [ ] I choose the right tool: branded types vs discriminated unions vs validation

### Runtime Bridge (Level 07)
- [ ] I can explain why `as` casts are dangerous at trust boundaries
- [ ] I use schema-first design (types derived from schemas, not the reverse)
- [ ] I validate at every trust boundary: API input, API output, env vars, DB results
- [ ] I can explain how Zod derives types from runtime schemas (phantom types + infer)
- [ ] I never use `JSON.parse(data) as T` without validation

### Library Patterns (Level 08)
- [ ] My public APIs require zero type annotations from consumers
- [ ] I know when to use overloads vs generics
- [ ] I use `any` intentionally inside implementations with safe public types
- [ ] I use `Prettify<T>` to make hover types readable
- [ ] I separate internal and public types in my libraries

### Compiler (Level 09)
- [ ] I know what every `strict` sub-flag does
- [ ] I enable `noUncheckedIndexedAccess` and `exactOptionalPropertyTypes`
- [ ] I can diagnose TypeScript performance issues with `--extendedDiagnostics`
- [ ] I use project references in monorepos
- [ ] I understand `moduleResolution` and when to use each option

### Debugging (Level 10)
- [ ] I read type errors bottom-up and can usually identify the root cause
- [ ] I break complex types into intermediate variables for debugging
- [ ] I can use `IsExact`, `IsAny`, `IsNever` for type-level assertions
- [ ] I know how to handle the `"excessively deep"` error
- [ ] I can debug distributive conditional type behavior

---

## Continued Learning

### Resources

| Resource | What it's for |
|----------|---------------|
| [type-challenges](https://github.com/type-challenges/type-challenges) | Practice type-level programming (puzzles) |
| [TypeScript Playground](https://www.typescriptlang.org/play) | Interactive type exploration |
| [Matt Pocock's Total TypeScript](https://totaltypescript.com/) | Video-based advanced TS |
| [TypeScript GitHub issues](https://github.com/microsoft/TypeScript/issues) | See how the team thinks about design decisions |
| [DefinitelyTyped](https://github.com/DefinitelyTyped/DefinitelyTyped) | Read real `.d.ts` files for complex libraries |
| [TypeScript Deep Dive (book)](https://basarat.gitbook.io/typescript) | Comprehensive reference |
| [Zod source code](https://github.com/colinhacks/zod) | World-class type system design example |
| [tRPC source code](https://github.com/trpc/trpc) | Generic inference chain masterclass |

### Type Challenge Recommendations

Start with these in order:

1. **Easy:** `Pick`, `Readonly`, `First of Array`, `Tuple to Object`
2. **Medium:** `Get Return Type`, `Omit`, `Deep Readonly`, `Chainable`
3. **Hard:** `Union to Intersection`, `String to Number`, `Join`, `Permutation`
4. **Extreme:** `Parse URL Params`, `JSON Parser`, `Type-level SQL`

### Books

| Book | Focus |
|------|-------|
| *Programming TypeScript* (Boris Cherny) | Advanced type system |
| *Effective TypeScript* (Dan Vanderkam) | Practical tips, 62 specific pieces of advice |
| *TypeScript in 50 Lessons* (Stefan Baumgartner) | Deep dives into specific features |

---

## Final Note

If you've worked through this entire curriculum and completed at least 2 capstone projects, you are in the top ~5% of TypeScript developers in terms of type system understanding.

The remaining work is application: use these skills in real codebases, real libraries, real APIs. The patterns become intuitive with repetition. The compiler becomes a tool, not an obstacle.

**You don't just use TypeScript. You think in it.**
