# draw-batching: One Draw Call via Primitive Buffers

## Guideline

Store the UI as compact, tightly-packed primitive definitions and let the vertex shader synthesize vertices from a metadata-carrying index buffer, so the whole UI submits in a single draw call with no state switches.

## Rationale

A naive IMGUI expands every rectangle into four padded vertices and pushes them over the bus each frame; with tens of thousands of rectangles that bandwidth and the per-state draw-call churn dominate. Storing one ~20-byte primitive instead of ~48 bytes of vertices, and encoding everything the GPU needs (primitive type, which corner, clip, texture) into the data, lets a single shader and a single draw call render the entire UI.

## How to Apply

1. Write primitives into a shared buffer with **no common stride** — a rectangle and a glyph can differ in size; pack each tightly.
2. Encode each index as 32 bits: low 24 bits = byte offset into the primitive buffer, top 8 bits = primitive type (6 bits) + corner id (2 bits). The vertex shader reads the primitive at that offset and synthesizes the corner vertex, bypassing the fixed-function input assembler.
3. Avoid extra draw calls: use **bindless textures** instead of atlasing, handle textured and untextured primitives in one shader, and **encode the clip rectangle's offset into each primitive** (CPU culls, GPU clips) rather than issuing scissor draws.
4. Render overlays (popups, menus) into a separate buffer, then append it to the main buffer at submit time, offsetting overlay indices by the main buffer's current index count.
5. Separate memory allocation from drawing: the caller pre-allocates buffer capacity; the drawing layer only writes. There is no automatic growth.

## Example

```c
// Index encoding: [type:6][corner:2][offset:24] -> shader builds the vertex from the primitive
uint32_t idx = (prim_type << 26) | (corner << 24) | byte_offset;

struct rect_t { float x, y, w, h; uint8_t col[4]; uint32_t clip_offset; }; // ~24B, one primitive
// vs the naive approach: 4 * {float x,y; uint8_t col[4];} padded vertices + 6 indices per rect
```

## Counter-Example

A UI with a handful of elements (a debug overlay, a splash) doesn't need primitive-buffer compaction — expanded vertices are fine. The technique earns its complexity when primitive counts reach the thousands.

## Related

[dpi-scaling.md](./dpi-scaling.md), [frame-delay.md](./frame-delay.md)
