# object-model: Typed Objects Described by Runtime Schemas

## Guideline

Represent every piece of tool/editor state as a typed object whose shape is described by a runtime type definition (a schema), not by a hand-written C struct, and store instance data as plain values keyed by property.

## Rationale

A central data model has to do generic work over every object — notify on change, undo, serialize, copy, diff — without knowing each type at compile time. If types are ordinary C structs, every one of those operations needs hand-written code per struct and breaks the moment a field is added. Describing types by data (a list of properties with kinds) lets one generic engine handle save/load, change tracking, and undo for all types, and lets new types or properties be introduced at runtime (loaded from a file, added by a plugin) without recompiling. The cost is one indirection per property access, which is negligible for editor-rate workloads.

## Techniques

- **Type definition** - A type is a name plus an ordered list of property definitions `{ name, kind, default }`. Register types into a registry keyed by a small type id.
- **Property kinds** - A closed set covers most models: `bool`, `int`, `float`, `string`, `reference` (id of another object), `sub_object` (owned child object), `buffer` (opaque bytes / blob), and `array` of any of the above.
- **Instance = values only** - An object instance carries its type id, its own id, and the property values; it does not carry layout or behavior. Two instances of the same type share one type definition.
- **Separate metadata from instance data** - The schema (names, kinds, defaults) lives once in the registry; instances store only values. Adding a property updates the schema; existing instances fall back to the default until set.
- **Data-driven extension** - Because types are data, a plugin or a loaded file can add a type or a property at runtime. Generic code keeps working; only code that reads that specific property needs to know about it.
- **Default + presence** - A property either has a set value or falls back to its default; track presence so "unset" is distinguishable from "set to the default" when that matters (e.g. for migration and diffing).

## How to Apply

1. Define a `property_kind` enum and a `property_def { name, kind, default }`.
2. Define a `type_def { name, prop_count, props[] }` and register it under a `type_id`.
3. Give each object an `id`, a `type_id`, and a value store indexed by property index.
4. Access properties generically: `get(obj, prop_index) -> value`, `set(obj, prop_index, value)`; never reach into a concrete struct.

## Example

```c
typedef enum {
  PROP_BOOL, PROP_INT, PROP_FLOAT, PROP_STRING,
  PROP_REFERENCE, PROP_SUB_OBJECT, PROP_BUFFER, PROP_ARRAY,
} property_kind_t;

typedef struct {
  const char     *name;
  property_kind_t kind;
  value_t         default_value; // used until the instance sets it
} property_def_t;

typedef struct {
  const char     *name;
  uint32_t        prop_count;
  property_def_t *props;
} type_def_t;

// Good: one generic getter works for every type, present or future.
static value_t obj_get(const object_t *o, uint32_t prop) {
  const type_def_t *t = type_registry_get(o->type_id);
  if (!value_is_set(&o->values[prop])) return t->props[prop].default_value;
  return o->values[prop];
}

// Bad: a concrete struct hard-codes the shape; serialize/undo/notify must be
// re-written per struct, and adding a field breaks every saved file.
typedef struct { float x, y; char *name; } node_t; // not data-driven
```

## Gotchas

- A generic value store costs an indirection per access — fine for editor edits, wrong for per-frame simulation iteration (that wants tight typed arrays instead).
- Property indices must stay stable for the lifetime of a session; renaming or reordering properties without a migration map invalidates instance data.
- Defaults are part of the schema: changing a default silently changes the meaning of every unset instance unless presence is tracked.

## Related

[references/references-and-ownership.md](./references-and-ownership.md), [references/change-notification.md](./change-notification.md), [references/serialization.md](./serialization.md)
