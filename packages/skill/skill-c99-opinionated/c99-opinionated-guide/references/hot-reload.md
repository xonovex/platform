# hot-reload: Hot-Reloadable Native Modules

**Guideline:** Make native code reloadable at runtime by putting the module behind a function-pointer API table and keeping all persistent state in host-owned memory — never in the module's globals/statics — so the shared library can be swapped without losing state.

**Rationale:** Reloading a `.so`/`.dll` while the program runs gives near-instant iteration (change code, rebuild the module, see it live — no restart, no lost session). It works only with discipline: when the old library is unloaded, everything that lived inside it vanishes — its static/global variables and every function pointer into it. So the contract is: the host owns the state and passes it in on every call (caller-owns-memory), and the host reaches the module only through an API struct it re-fetches after each reload. Code that caches a module function pointer or stashes state in a module global breaks the moment you reload.

**How to Apply:**

1. Define an API struct of function pointers the module fills in; the host calls only through it.
2. Keep all persistent state in a host-owned struct passed into every entry point — the module is stateless across calls.
3. Build the module as a shared library exporting one symbol (e.g. `load_module`) that returns the populated API.
4. Host loads the library (`dlopen`/`LoadLibrary`), fetches `load_module`, and calls through the returned table.
5. Watch the file's mtime; on change, unload the old library, load the new one, re-fetch the API. State survives because it never lived in the module.
6. Re-read the API table after every reload; never cache a module function pointer across a frame boundary.

**Example:**

```c
// shared module API — host calls only through this table
typedef struct game_state game_state;          // defined/owned by the HOST
typedef struct {
    void (*update)(game_state *s, float dt);    // no state inside the module
    void (*render)(game_state *s);
} module_api;

// in the module (built as libgame.so): the one exported entry point
const module_api *load_module(void) {
    static const module_api api = { .update = game_update, .render = game_render };
    return &api;                                // functions live in the .so; re-fetch after reload
}

// in the host
void *lib = dlopen("./libgame.so", RTLD_NOW);
const module_api *api = ((const module_api *(*)(void))dlsym(lib, "load_module"))();
// each frame: api->update(state, dt); api->render(state);
// on file change: dlclose(lib); lib = dlopen(...); api = ...;  // `state` persists (host-owned)
```

**Gotchas:**

- Any `static`/global variable inside the module is reset (or lost) on reload — keep mutable state host-side; module statics are for immutable tables only.
- A function pointer into the module dangles after `dlclose`; always indirect through the freshly re-fetched API table, never a cached pointer.
- Changing the layout of the host-owned state struct across a reload corrupts live state — version it and migrate, or accept that layout changes need a restart.
- Reload at a known-safe point (e.g. top of the frame), not mid-callback; a reload while module code is on the stack is a use-after-free.
- On Windows the loaded DLL is locked — load a copy (and handle the `.pdb` lock) so the build can overwrite the original.

**Related:** [references/caller-owns-memory.md](./caller-owns-memory.md), [references/composability.md](./composability.md), [references/file-naming.md](./file-naming.md)
