# Getting Started

## Quick Verification

Before starting the curriculum, verify your environment:

```bash
# Node 25+ with native TypeScript
node --version   # Should be v25.x.x or higher
node -e "const x: number = 1; console.log('TS works:', x)"

# Bun
bun --version

# System tools
which strace 2>/dev/null || which dtrace 2>/dev/null
which perf 2>/dev/null || echo "Install linux-tools for perf"
```

## First Exercise

Create a file called `runtime-check.ts`:

```typescript
// runtime-check.ts — Verify your environment
const nodeVersion = process.versions.node;
const v8Version = process.versions.v8;
const uvVersion = process.versions.uv;

console.log(`Node.js: v${nodeVersion}`);
console.log(`V8:      v${v8Version}`);
console.log(`libuv:   v${uvVersion}`);
console.log(`Platform: ${process.platform}`);
console.log(`Arch:     ${process.arch}`);
console.log(`PID:      ${process.pid}`);
console.log(`Memory:   ${JSON.stringify(process.memoryUsage(), null, 2)}`);
console.log(`CPUs:     ${require("os").cpus().length}`);
```

Run it:

```bash
node runtime-check.ts
```

If all values print correctly, you're ready. Proceed to [Module 01 — Runtime Architecture](01-runtime-architecture/README.md).
