# target-types: Target Types

**Guideline:** Choose appropriate target types based on how the code will be consumed.

**Rationale:** Different target types optimize for different use cases: shared libraries for runtime linking, static for embedding, header-only for templates, object for reusing compiled code.

**Example:**

```cmake
# Normal library (static or shared based on BUILD_SHARED_LIBS)
add_library(mylib src/lib.cpp)

# Explicit static library
add_library(static_lib STATIC src/static.cpp)

# Explicit shared library
add_library(shared_lib SHARED src/shared.cpp)

# Header-only library
add_library(header_only INTERFACE)
target_include_directories(header_only INTERFACE include)

# Object library for intermediate compilation
add_library(objects OBJECT src/common.cpp)
target_link_libraries(mylib PRIVATE objects)

# Executable and GUI app
add_executable(app src/main.cpp)
add_executable(gui WIN32 MACOSX_BUNDLE src/gui.cpp)
```

**Techniques:**

- add_library() without type: Let BUILD_SHARED_LIBS control static vs shared
- STATIC: Always statically link, good for embedded or single-binary scenarios
- SHARED: Dynamic library for runtime linking and flexible updates
- INTERFACE: Header-only library with no compiled code
- OBJECT: Intermediate object library for sharing compiled units across targets
