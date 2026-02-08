# testing: Testing with CTest

**Guideline:** Use CTest to integrate automated testing into the build system.

**Rationale:** CTest provides a standardized way to run tests across platforms with features like timeouts, labels, and parallel execution.

**Example:**

```cmake
# Enable testing
enable_testing()

# Add test executable
add_executable(unit_tests tests/test_core.cpp tests/test_utils.cpp)
target_link_libraries(unit_tests PRIVATE core gtest_main)

# Register test cases
add_test(NAME unit_tests COMMAND unit_tests)
add_test(NAME integration_test COMMAND app --test-mode WORKING_DIRECTORY ${CMAKE_BINARY_DIR})

# Set test properties
set_tests_properties(unit_tests PROPERTIES TIMEOUT 30 LABELS "unit")
```

**Techniques:**

- enable_testing(): Enable the testing feature in CMake project
- add_test(): Register test executable or command with optional working directory
- set_tests_properties(): Set timeout, labels, or other test attributes
- add_executable() for tests: Build test binaries separately from production code
- LABELS: Tag tests for selective running with ctest -L filter
