# mini-notation: Pattern Mini-Notation Syntax

Express rhythmic and melodic patterns with mini-notation operators inside `s()`/`note()` strings.

| Operator  | Syntax    | Meaning                    |
| --------- | --------- | -------------------------- |
| Space     | `a b c`   | Sequential steps           |
| Rest      | `~`       | Silence                    |
| Multiply  | `a*4`     | Repeat 4x within the step  |
| Divide    | `a/2`     | Stretch over 2 cycles      |
| Subdivide | `[a b]`   | Fit both into one step     |
| Alternate | `<a b c>` | One value per cycle        |
| Elongate  | `a@2`     | Double the step's duration |
| Replicate | `a!3`     | Repeat as 3 separate steps |
| Parallel  | `a,b`     | Simultaneous layers        |
| Euclidean | `a(3,8)`  | 3 hits spread over 8 steps |

```javascript
s("bd*4 hh*8"); // 4 kicks, 8 hi-hats in one cycle
s("[bd [sd sd]] [hh hh hh hh]"); // nested subdivisions
s("bd(3,8)"); // euclidean rhythm
s("bd*4, hh*16"); // parallel kick + hat layers
note("<c3 e3 g3 b3>"); // one note per cycle
```
