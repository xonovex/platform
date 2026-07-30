# testing: Testing with CTest

Pass `WORKING_DIRECTORY` on `add_test` when a test must run from a specific directory.

```cmake
enable_testing()
add_test(NAME integration_test COMMAND app --test-mode WORKING_DIRECTORY ${CMAKE_BINARY_DIR})
```
