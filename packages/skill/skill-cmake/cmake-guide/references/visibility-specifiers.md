# visibility-specifiers: Visibility Specifiers

`PUBLIC` = target and consumers; `PRIVATE` = target only; `INTERFACE` = consumers only. Apply the same keyword consistently across includes/definitions/options/links, and minimize `PUBLIC` to shrink transitive dependencies.

```cmake
target_include_directories(mylib PUBLIC include)         # both
target_include_directories(mylib PRIVATE src/internal)   # target only
add_library(header_only INTERFACE)
target_include_directories(header_only INTERFACE include) # consumers only
target_link_libraries(app PRIVATE mylib PUBLIC common INTERFACE iface)
```
