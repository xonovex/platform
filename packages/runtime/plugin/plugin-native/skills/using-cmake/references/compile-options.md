# compile-options: Compile Definitions

Set preprocessor macros on targets with `target_compile_definitions` and explicit visibility.

```cmake
target_compile_definitions(mylib
    PUBLIC API_VERSION=2
    PRIVATE $<$<CONFIG:Debug>:DEBUG_BUILD>)
```
