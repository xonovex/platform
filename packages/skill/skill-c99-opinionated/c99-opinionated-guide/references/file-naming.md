# file-naming: File Naming Conventions

**Guideline:** Suffixes for purpose: `_type.h` (types), `_impl.h` (scalar), `_aos/_soa/_simd` (variants).

**Rationale:** Enables quick file identification, prevents include order issues, supports multiple variant implementations.

**Example:**

```
vector_type.h        # Types only
vector_impl.h        # Scalar implementations
vector.h             # Includes _type + _impl
vector_aos.h         # Batch AoS
vector_soa.h         # Batch SoA (cache-friendly)
vector_simd_impl.h   # SIMD scalar
vector_soa_simd.h    # SIMD SoA batch
vector.test.c        # Scalar tests

// vector.h
#include "vector_type.h"
#include "vector_impl.h"

// vector_soa_simd.h
#include "vector_type.h"
#include "vector_soa.h"
```

**Techniques:**

- Type suffix: Use `_type.h` for type definitions only
- Impl suffix: Use `_impl.h` for inline scalar implementations
- Batch variants: Use `_aos.h`, `_soa.h` for batch implementations
- SIMD suffix: Use `_simd.h` for SIMD implementations
- Dimension modules: Use `{name}{dim}.h` pattern like `aabb2d.h`, `sphere3d_simd.h`
