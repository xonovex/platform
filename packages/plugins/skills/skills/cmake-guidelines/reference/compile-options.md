# compile-options: Compile Options and Definitions

**Guideline:** Set compiler flags, definitions, and features on targets using target*compile*\* commands with appropriate visibility.

**Rationale:** Target-scoped settings prevent pollution of the global build environment and allow different targets to have different requirements.

**Example:**

```cmake
# Set compile options
target_compile_options(mylib
    PRIVATE
        $<$<CXX_COMPILER_ID:GNU,Clang>:-Wall -Wextra -pedantic>
        $<$<CXX_COMPILER_ID:MSVC>:/W4>
)

# Set compile definitions
target_compile_definitions(mylib
    PUBLIC API_VERSION=2
    PRIVATE
        $<$<CONFIG:Debug>:DEBUG_BUILD>
        $<$<CONFIG:Release>:RELEASE_BUILD>
)

# Set compile features
target_compile_features(mylib PUBLIC cxx_std_20)
```

**Techniques:**

- target_compile_options(): Set compiler-specific flags with generator expressions
- target_compile_definitions(): Define preprocessor macros with visibility control
- target_compile_features(): Require C++ standard using cxx_std_20, cxx_std_17, etc
- Compiler detection: Use $<$<CXX_COMPILER_ID:GNU,Clang>:flags> for portability
- Configuration-based: Apply $<$<CONFIG:Debug>:DEBUG> for build-type-specific settings
