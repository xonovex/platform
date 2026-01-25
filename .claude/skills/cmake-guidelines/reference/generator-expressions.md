# generator-expressions: Generator Expressions

**Guideline:** Use generator expressions for conditional and context-dependent CMake settings.

**Rationale:** Generator expressions evaluate at build time, allowing different settings based on compiler, platform, configuration, or build/install context.

**Example:**

```cmake
# Build/Install interface paths
target_include_directories(mylib PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)

# Compiler-specific options
target_compile_options(mylib PRIVATE
    $<$<CXX_COMPILER_ID:GNU>:-fno-rtti>
    $<$<CXX_COMPILER_ID:MSVC>:/GR->
)

# Configuration and platform-specific settings
target_compile_definitions(mylib PRIVATE
    $<$<CONFIG:Debug>:DEBUG_MODE=1>
    $<$<PLATFORM_ID:Linux>:LINUX_BUILD>
)
```

**Techniques:**
- BUILD_INTERFACE/INSTALL_INTERFACE: Different paths for build vs installed consumption
- CXX_COMPILER_ID: Detect compiler and apply vendor-specific flags
- CONFIG: Apply Debug or Release build-specific settings
- PLATFORM_ID: Apply Linux, Windows, Darwin-specific settings
- Nested conditions: Chain $<$<X>:$<Y:value>>> for complex logic
