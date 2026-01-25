# installation: Installation

**Guideline:** Configure installation rules to enable package distribution and consumption via find_package.

**Rationale:** Proper installation allows your library to be installed system-wide or in custom locations and discovered by other CMake projects.

**Example:**

```cmake
# Install targets with export
install(TARGETS mylib app
    EXPORT mylib-targets
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
    INCLUDES DESTINATION include
)

# Install headers
install(DIRECTORY include/
    DESTINATION include
    FILES_MATCHING PATTERN "*.h"
)

# Export targets for find_package discovery
install(EXPORT mylib-targets
    FILE mylib-targets.cmake
    NAMESPACE mylib::
    DESTINATION lib/cmake/mylib
)
```

**Techniques:**
- install(TARGETS): Install libraries and executables with type-specific destinations
- EXPORT: Create importable CMake files for downstream find_package() calls
- DESTINATION: Set lib/bin/include paths for each artifact type
- NAMESPACE: Prefix exported targets (mylib::) to prevent naming conflicts
- install(DIRECTORY): Install headers with glob patterns for selective inclusion
