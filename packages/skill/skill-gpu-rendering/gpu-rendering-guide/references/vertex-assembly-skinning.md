# vertex-assembly-skinning: Programmable Vertex Fetch and GPU Skinning

## Guideline

Fetch vertex data yourself from storage buffers in the shader (programmable vertex pull) behind a small loader interface, instead of binding fixed-function vertex input; this lets one shader read any packing, any channel set, and any vertex — which is what makes flexible GPU skinning and morph targets fall out cleanly.

## Rationale

The fixed-function input assembler ties a pipeline to one rigid vertex layout and only ever hands the shader _its own_ vertex. Pulling vertices manually from byte-addressable storage buffers removes both limits: a single shader supports many packings/quantizations by going through an abstract loader (`load_position()`, `load_normal()`, …) that hides offsets and strides; it can read _any_ vertex or the index buffer from any stage; and it can gracefully skip channels a mesh doesn't have. Skinning then needs no special vertex format — the shader just also reads bone influences and matrices from buffers and blends, and variable bone-per-vertex counts cost no wasted space because the influence list is indirected, not padded to a fixed maximum.

## How to Apply

1. Store vertex streams in storage/byte-address buffers; describe each mesh with active-channel bitflags plus per-channel offset and stride, and a vertex count.
2. Expose a loader context the shader calls (`load_position(i)`, `load_texcoord0(i)`, …) so material shaders are independent of packing; return sensible defaults for absent channels.
3. For skinning, pack each vertex's influence reference as one word — a count plus an offset into a shared influence buffer of `{bone_index, weight}` pairs — instead of fixed N bones per vertex.
4. Linear-blend skin in the vertex shader: loop the influences, accumulate `weight * bone_matrix * position` (and the matrix's rotation for normals/tangents).
5. Ping-pong the bone-matrix buffer between frames (current + previous) so you can output motion vectors from skinned previous positions.
6. Let skinning register as a named shader system that rendering discovers after culling, so it composes with materials without hard coupling.

## Example

```glsl
// Programmable pull + linear blend skinning in the vertex shader.
// Buffers (bindless or bound): vertex streams, influences, bone matrices.
struct Influence { uint bone; float weight; };

vec3 skin_position(uint vtx) {
    vec3  p     = load_position(vtx);          // loader hides packing/offset/stride
    uint  skin  = load_skin_data(vtx);         // hi 8 bits = count, lo 24 = offset
    uint  count = skin >> 24, base = skin & 0xFFFFFFu;
    vec3  acc   = vec3(0.0);
    for (uint i = 0u; i < count; ++i) {
        Influence inf = influences[base + i];
        acc += inf.weight * (bone_matrices[inf.bone] * vec4(p, 1.0)).xyz;
    }
    return acc;                                // normals/tangents use the 3x3 part
}
```

## Gotchas

- Manual fetch bypasses any format conversion the input assembler would do — you must decode/normalize packed/quantized data yourself in the loader.
- Forgetting to skin the normal/tangent (only the position) breaks lighting on animated meshes; skin the basis with the matrix's rotation part.
- A fixed bone-count vertex format wastes memory on low-influence vertices and clips high-influence ones; indirect the influence list.
- Motion vectors need _previous-frame_ skinned positions — keep last frame's bone matrices, don't recompute from current.
- Compute-skinning once into a buffer (vs re-skinning in every pass) pays off when a mesh is drawn many times per frame (shadow + depth + color); skinning in the vertex shader re-does the work per pass.

## Related

[references/shader-system.md](./shader-system.md), [references/binding-model.md](./binding-model.md), [references/gpu-compute-simulation.md](./gpu-compute-simulation.md), **c99-game-opinionated-guide**
