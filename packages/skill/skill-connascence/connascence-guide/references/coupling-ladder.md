# The Coupling Ladder

A worst-to-best ranking of seams so a reviewer can grade any seam by name and push it down a rung.

## How to grade

- **Replace informal "tight/loose" with a rung name** — point at the call site, name the worst rung it sits on, and move it one rung down. The ranking is total: each rung is strictly preferable to the one above it.
- **One target per seam** — the lower the rung, the more local and the easier to change in isolation. Data coupling is the floor; aim there.
- **Pair with connascence** — the rung names the _shape_ of the dependency; connascence names the _forced co-change_. Grade both, then act. See [connascence.md](connascence.md).

## The ladder (worst to best)

- **Content (worst)** — one module reaches into another's internals: its fields, its private state, its representation. The two now change together with no surface mediating them.
- **Common** — modules share global or mutable state. Every reader and writer is silently bound through the shared cell; an edit anywhere can break anyone.
- **Control** — a caller passes a flag that drives the callee's branching. The caller now knows the callee's internal control flow and must change with it.
- **Stamp** — a caller passes a whole record but the callee uses one field. The signature over-declares what it depends on, dragging the whole type along for one value.
- **Data (best)** — plain parameters carrying no control. The callee depends only on the values it actually reads.

## One BAD -> GOOD per rung

Content — reach into a field vs ask a method:

```go
// BAD — billing mutates the Order's private representation.
order.total = order.total + tax
// GOOD — ask the owner; depend on one surface.
order.AddCharge(tax)
```

Common — a package global vs a passed-in value:

```go
// BAD — two call sites mutate one shared cell; order and reach are invisible.
var activeCart *Cart
func addItem() { activeCart.Add(item) }
// GOOD — the composition root passes the cart in explicitly.
func addItem(cart *Cart) { cart.Add(item) }
```

Control — a flag that forks the callee vs two methods:

```go
// BAD — caller drives the callee's branch with a flag.
repo.Save(order, dryRun) // callee: if dryRun { validate } else { persist }
// GOOD — split the branch into named operations.
repo.Validate(order)
repo.Persist(order)
```

Stamp — pass the whole record to read one field vs pass the field:

```go
// BAD — Discount only needs the subtotal, but takes the whole Order.
func Discount(order *Order) int { return order.Subtotal / 10 }
// GOOD — depend on exactly the value read.
func Discount(subtotal int) int { return subtotal / 10 }
```

Data — an opaque opts map vs explicit params:

```go
// BAD — opts smuggles untyped control and hides what is read.
func Price(opts map[string]any) {}
// GOOD — explicit values, no control channel.
func Price(item *Item, qty int) {}
```

## Tie to connascence

- **Content and common are the strongest, least-local rungs** — they bind through a representation or a shared cell that neither call site shows, so they carry strong, non-local connascence (often of value or meaning). Treat them as the first to fix.
- **Control coupling is connascence of meaning crossing a boundary** — both sides must agree on what the flag means. Splitting the flag into named operations converts it toward connascence of name.
- **Stamp and data are the weakest rungs** — a data seam carries only connascence of name and type on the values it names, which is the most local, compile-time-visible form. That is why pushing down the ladder and weakening connascence are the same move.

## Worked grade

- **Seam** — a billing function reads an Order's private `total` field to decide whether to apply a surcharge. That is content coupling plus connascence of meaning across a boundary: billing depends on the Order's representation and on a silent pact about the field.
- **Fix** — have the Order self-declare a neutral `Total()` value the billing function reads. The seam drops to data coupling, the agreement becomes named, and billing stays correct against order types it has never seen.

```go
// GOOD — the Order self-declares; billing reads a neutral value.
if order.Total() >= threshold { return nil }
```

A train wreck like `order.customer().address().city()` is the call-site tell for the worst rung: it reaches through two objects the caller never received. Use the [law of demeter](law-of-demeter.md) as the content-coupling detector.

Back to the overview: [SKILL.md](../SKILL.md).
