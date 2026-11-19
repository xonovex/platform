# installation: Installation

Install targets under an `EXPORT` set, then install that export as a `*-targets.cmake` with a `NAMESPACE` so downstream `find_package` gets namespaced imported targets. Pair `INSTALL_INTERFACE` include dirs (see generator-expressions) with `INCLUDES DESTINATION`.

```cmake
install(TARGETS mylib app
    EXPORT mylib-targets
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
    RUNTIME DESTINATION bin
    INCLUDES DESTINATION include)
install(DIRECTORY include/ DESTINATION include FILES_MATCHING PATTERN "*.h")
install(EXPORT mylib-targets
    FILE mylib-targets.cmake
    NAMESPACE mylib::
    DESTINATION lib/cmake/mylib)
```
