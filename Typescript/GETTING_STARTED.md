# Getting Started

## Environment Setup

### TypeScript Installation

```bash
# Global (for playground scripts)
npm install -g typescript ts-node

# Verify
tsc --version  # Should be 5.0+
```

### Recommended: TypeScript 5.4+

This curriculum uses features available in TypeScript 5.4+. Key features by version:

| Version | Feature Used |
|---------|-------------|
| 4.1 | Template literal types, key remapping |
| 4.5 | Tail-call recursion for conditional types |
| 4.7 | `extends` constraints on `infer` |
| 4.9 | `satisfies` operator |
| 5.0 | `const` type parameters |
| 5.2 | `using` declarations |
| 5.4 | Narrowing in closures |

### Editor Setup

**VS Code** with these extensions:
- TypeScript built-in (ships with VS Code)
- [Error Lens](https://marketplace.visualstudio.com/items?itemName=usernamehw.errorlens) — inline type errors
- [Pretty TypeScript Errors](https://marketplace.visualstudio.com/items?itemName=yoavbls.pretty-ts-errors) — readable error messages
- [Twoslash Query Comments](https://marketplace.visualstudio.com/items?itemName=Orta.vscode-twoslash-queries) — `// ^?` hover inspection

### Settings

Add to your VS Code `settings.json`:

```json
{
  "typescript.tsdk": "node_modules/typescript/lib",
  "typescript.enablePromptUseWorkspaceTsdk": true,
  "editor.inlayHints.enabled": "on",
  "typescript.inlayHints.parameterNames.enabled": "all",
  "typescript.inlayHints.variableTypes.enabled": true,
  "typescript.inlayHints.functionLikeReturnTypes.enabled": true
}
```

These inlay hints show you what TypeScript is **inferring** — essential for understanding inference behavior.

---

## Project Setup (For Exercises)

```bash
mkdir ts-deep-dive && cd ts-deep-dive
npm init -y
npm install -D typescript @types/node
npx tsc --init
```

### Recommended `tsconfig.json` for this curriculum:

```json
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "NodeNext",
    "moduleResolution": "NodeNext",
    "strict": true,
    "exactOptionalPropertyTypes": true,
    "noUncheckedIndexedAccess": true,
    "noPropertyAccessFromIndexSignature": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "declaration": true,
    "declarationMap": true,
    "sourceMap": true,
    "outDir": "./dist",
    "rootDir": "./src"
  },
  "include": ["src"]
}
```

**Why these flags?** Explained in detail in [Level 09](09-compiler-tooling/).

---

## How to Run Examples

### Option 1: TypeScript Playground (Recommended for type-only exploration)

https://www.typescriptlang.org/play

Paste code → hover over types → modify → observe.

### Option 2: Local file with `ts-node`

```bash
npx ts-node src/example.ts
```

### Option 3: `tsc --watch` for compile-time experiments

```bash
npx tsc --watch --noEmit
```

This gives you a live feedback loop. Write types, save, see errors.

---

## Mental Setup

Before you begin, internalize this:

1. **You are learning a second language.** TypeScript has two evaluation phases — runtime (JavaScript) and compile-time (the type system). The type system has its own variables (generics), conditionals (`extends ? :`), loops (mapped types), and functions (generic type aliases).

2. **Hover is your REPL.** In VS Code, hovering over a type expression shows you what it resolves to. This is your primary debugging tool.

3. **Errors are data.** Don't just read the first line of a type error. Read the entire message. It tells you exactly what the compiler tried and where it failed.

4. **Break the code.** Every example in this curriculum should be modified. Change a type. Remove a constraint. See what breaks and *why*.
