# visibility-specifiers: Visibility Specifiers

**Guideline:** Use PUBLIC, PRIVATE, and INTERFACE to control how target properties propagate to consumers.

**Rationale:** Proper visibility prevents leaking implementation details and ensures consumers get only the dependencies they need, reducing coupling and build times.

**Example:**

```cmake
# PUBLIC: Library AND consumers need this
target_include_directories(mylib PUBLIC include)
target_compile_definitions(mylib PUBLIC USE_FEATURE=1)

# PRIVATE: Only library needs this
target_include_directories(mylib PRIVATE src/internal)
target_compile_definitions(mylib PRIVATE DEBUG_MODE=1)

# INTERFACE: Only consumers need this
add_library(header_only INTERFACE)
target_include_directories(header_only INTERFACE include)

# Linking with visibility
add_executable(app src/main.cpp)
target_link_libraries(app PRIVATE mylib PUBLIC common INTERFACE interface)
```

**Techniques:**

- PUBLIC: Properties visible to both target and all downstream consumers
- PRIVATE: Properties used only by target, not propagated to consumers
- INTERFACE: Properties for consumers only, not used by target itself
- Consistent visibility: Apply same visibility to includes, links, definitions, options
- Minimize PUBLIC: Reduce PUBLIC properties to shrink transitive dependencies
