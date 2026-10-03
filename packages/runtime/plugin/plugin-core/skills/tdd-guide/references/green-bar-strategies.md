# tdd: Going Green: Obvious Implementation, Fake It, Triangulation

Three ways to make a red test pass; choose by how sure you are of the real code.

- **Obvious Implementation**: you know the real code and it's small: type it. This is the default; most steps are obvious implementations.
- **Fake It**: unsure of the real code: return a constant that satisfies the test. Now the same literal lives in the test and the code; removing that duplication during refactor is what drives you to the general implementation. Fake It turns "I don't know how to write this" into "I know how to remove this duplication".
- **Triangulation**: generalize only once two or more concrete examples force it: fake one test, add a second with a different value the fake can't satisfy, abstract to the shape both pin down.

```
Is the real implementation obvious right now?
  yes -> Obvious Implementation
  no  -> Fake It (return a constant; go green)
           Can you now see how to remove the test<->code duplication?
             yes -> generalize directly during refactor
             no  -> Triangulate (second example forces the general form)
                      still stuck -> step away; the design question is bigger than the next line
```

## Cross-references

- The cycle these live inside, and why their duplication must be refactored away, see [red-green-refactor.md](red-green-refactor.md).
- Choosing the next test and letting the API emerge, see [test-list-and-design.md](test-list-and-design.md).
- Structuring the individual test (AAA): **testing-guide**.
