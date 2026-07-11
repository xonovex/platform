# compile-options: Compile Options and Definitions

Set flags, macros, and standard on targets via `target_compile_options` / `target_compile_definitions` / `target_compile_features` with explicit visibility — never global `add_compile_options`. Guard vendor flags with `$<$<CXX_COMPILER_ID:GNU,Clang>:...>`; require the standard with `cxx_std_20` rather than setting `CMAKE_CXX_STANDARD`.

```cmake
target_compile_options(mylib PRIVATE
    $<$<CXX_COMPILER_ID:GNU,Clang>:-Wall -Wextra -pedantic>
    $<$<CXX_COMPILER_ID:MSVC>:/W4>)
target_compile_definitions(mylib
    PUBLIC API_VERSION=2
    PRIVATE $<$<CONFIG:Debug>:DEBUG_BUILD>)
target_compile_features(mylib PUBLIC cxx_std_20)
```
