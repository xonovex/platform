# project-structure: Basic Project Structure

**Guideline:** Set up CMake projects with modern target-based configuration using CMake 3.20 or higher.

**Rationale:** A well-structured CMakeLists.txt establishes build requirements, project metadata, and target definitions that form the foundation for maintainable builds.

**Example:**

```cmake
cmake_minimum_required(VERSION 3.20)
project(demo VERSION 1.0.0 DESCRIPTION "Example" LANGUAGES C CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Build library
add_library(core src/core.cpp src/utils.cpp)
target_include_directories(core PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>
)
target_compile_features(core PUBLIC cxx_std_20)

# Build executable
add_executable(app src/main.cpp)
target_link_libraries(app PRIVATE core)

# Testing
enable_testing()
add_executable(core_test tests/core_test.cpp)
target_link_libraries(core_test PRIVATE core)
add_test(NAME core_test COMMAND core_test)
```

**Techniques:**
- cmake_minimum_required(): Set 3.20+ for modern CMake features
- project(): Declare name, version, description, and required languages
- add_library/add_executable(): Create targets for libraries and binaries
- target_include_directories(): Configure include paths with generator expressions
- target_compile_features(): Require specific C++ standard like cxx_std_20
