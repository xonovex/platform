# numeric-separator-enforcement: Underscores in Large Numeric Literals

Add `_` separators grouped by thousands to numeric literals `>= 10_000`; leave smaller numbers plain. Caught by ESLint `unicorn/numeric-separators-style`.

```typescript
const timeout = 30_000; // ✅
const maxUsers = 1_000_000; // ✅ grouped by thousands, not 100_0000
const port = 3000; // ✅ < 10_000, no separator
```
