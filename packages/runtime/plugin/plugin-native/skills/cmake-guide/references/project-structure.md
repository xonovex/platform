# project-structure: Basic Project Structure

Open with `cmake_minimum_required(VERSION 3.20)` then `project(... LANGUAGES ...)`. Define targets and wire include dirs with build/install interface expressions.

```cmake
cmake_minimum_required(VERSION 3.20)
project(demo VERSION 1.0.0 DESCRIPTION "Example" LANGUAGES C CXX)

add_library(core src/core.cpp src/utils.cpp)
target_include_directories(core PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>)
target_compile_features(core PUBLIC cxx_std_20)

add_executable(app src/main.cpp)
target_link_libraries(app PRIVATE core)

enable_testing()
add_executable(core_test tests/core_test.cpp)
target_link_libraries(core_test PRIVATE core)
add_test(NAME core_test COMMAND core_test)
```
