# mini-notation: Pattern Mini-Notation Syntax

**Guideline:** Use Strudel mini-notation operators to express rhythmic and melodic patterns concisely.

**Rationale:** Mini-notation provides compact syntax for complex structures without verbose function calls.

**Example:**

```javascript
// Simple drum pattern
s("bd sd ~ sd"); // Kick, snare, rest, snare

// Repeat with multiply
s("bd*4 hh*8"); // 4 kicks, 8 hi-hats in one cycle

// Subdivisions
s("[bd sd] [hh hh]"); // Two groups fitted in one step each

// Euclidean rhythm (3 hits over 8 steps)
s("bd(3,8)");

// Parallel layers
s("bd*4, hh*16"); // Kick layer + hi-hat layer simultaneously

// Chord progression
note("<c3 e3 g3 b3>"); // One note per cycle (ascending pattern)
```

## Core Operators

| Operator  | Syntax    | Description           |
| --------- | --------- | --------------------- |
| Space     | `a b c`   | Sequential            |
| Rest      | `~`       | Silence               |
| Multiply  | `a*4`     | Repeat 4x             |
| Divide    | `a/2`     | Stretch over 2 cycles |
| Subdivide | `[a b]`   | Fit both in one step  |
| Alternate | `<a b c>` | One per cycle         |
| Elongate  | `a@2`     | Double duration       |
| Replicate | `a!3`     | Repeat 3x             |
| Parallel  | `a,b`     | Simultaneous          |
| Euclidean | `a(3,8)`  | 3 hits over 8         |

## Common Patterns

```javascript
s("bd*4"); // 4 kicks per cycle (trance)
s("~ hh ~ hh"); // Offbeat hi-hats
s("~ sd ~ sd"); // Backbeat snare
s("[bd sd]*2"); // Kick-snare groups
s("bd*4, hh*8"); // Parallel layers
s("bd(3,8)"); // Euclidean rhythm
note("<c3 e3 g3 b3>"); // Chord progression per cycle
```

## Nesting and Complexity

```javascript
s("[bd [sd sd]] [hh hh hh hh]"); // Complex nesting
note("[c3 e3] [g3 b3]"); // Melodic subdivisions
s("[bd sd]*2, hh*16"); // Layered rhythm
```

**Techniques:**

- Space separator: Sequential patterns (a b c)
- Tilde (~): Rest/silence in pattern
- Multiply (*): Repeat pattern n times (a*4)
- Divide (/): Stretch pattern over n cycles
- Brackets []: Subdivide into tighter steps
- Angle brackets <>: Alternate one per cycle
- @ symbol: Elongate duration (a@2 doubles)
- ! symbol: Replicate pattern n times (a!3)
- Comma (,): Parallel/simultaneous patterns
- Euclidean notation: a(hits, steps) for rhythmic counting
