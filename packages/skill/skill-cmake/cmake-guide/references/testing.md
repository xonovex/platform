# testing: Testing with CTest

`enable_testing()`, build test binaries as separate executables, register with `add_test(NAME ... COMMAND ...)` (optionally `WORKING_DIRECTORY`), and set `TIMEOUT`/`LABELS` via `set_tests_properties` — filter labels with `ctest -L`.

```cmake
enable_testing()
add_executable(unit_tests tests/test_core.cpp tests/test_utils.cpp)
target_link_libraries(unit_tests PRIVATE core gtest_main)
add_test(NAME unit_tests COMMAND unit_tests)
add_test(NAME integration_test COMMAND app --test-mode WORKING_DIRECTORY ${CMAKE_BINARY_DIR})
set_tests_properties(unit_tests PROPERTIES TIMEOUT 30 LABELS "unit")
```
