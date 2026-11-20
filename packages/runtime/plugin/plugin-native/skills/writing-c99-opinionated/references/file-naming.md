# file-naming: File Naming Conventions

Suffix headers by purpose so variants and include order are unambiguous: `_type.h` (types only), `_impl.h` (inline scalar), `_aos.h`/`_soa.h`/`_simd.h` (batch/SIMD variants). Dimension modules use `{name}{dim}` like `aabb2d.h`, `sphere3d_simd.h`.

```
vector_type.h        # Types only
vector_impl.h        # Scalar implementations
vector.h             # Includes _type + _impl
vector_aos.h         # Batch AoS
vector_soa.h         # Batch SoA (cache-friendly)
vector_simd_impl.h   # SIMD scalar
vector_soa_simd.h    # SIMD SoA batch
vector.test.c        # Scalar tests
```
