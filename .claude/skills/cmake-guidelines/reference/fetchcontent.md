# fetchcontent: Dependency Management with FetchContent

**Guideline:** Use FetchContent to fetch and build external dependencies from source at configure time.

**Rationale:** FetchContent ensures dependencies are built with the same compiler and settings as your project, providing reproducible builds without system installation.

**Example:**

```cmake
include(FetchContent)

# Fetch external dependency
FetchContent_Declare(
    fmt
    GIT_REPOSITORY https://github.com/fmtlib/fmt.git
    GIT_TAG 10.1.1
    GIT_SHALLOW TRUE
)

FetchContent_MakeAvailable(fmt)

# Use the dependency
add_executable(app src/main.cpp)
target_link_libraries(app PRIVATE fmt::fmt)
```

**Techniques:**
- FetchContent_Declare(): Declare dependency with git repo and version pin
- GIT_TAG: Pin exact version for reproducible builds
- GIT_SHALLOW: Use shallow clone for faster fetching
- FetchContent_MakeAvailable(): Fetch and populate dependency with one call
- Populate: Use FetchContent_GetProperties() for conditional population control
