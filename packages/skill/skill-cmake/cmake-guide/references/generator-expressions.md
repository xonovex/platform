# generator-expressions: Generator Expressions

Use `$<...>` for context-dependent settings evaluated at build time, not configure time (so they never expand in `message()`). Key forms: `$<BUILD_INTERFACE:>`/`$<INSTALL_INTERFACE:>` split include paths for build vs installed use, `$<CONFIG:Debug>`, `$<PLATFORM_ID:Linux>`.

```cmake
target_include_directories(mylib PUBLIC
    $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    $<INSTALL_INTERFACE:include>)
target_compile_definitions(mylib PRIVATE
    $<$<CONFIG:Debug>:DEBUG_MODE=1>
    $<$<PLATFORM_ID:Linux>:LINUX_BUILD>)
```
