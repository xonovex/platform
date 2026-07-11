# fetchcontent: Dependency Management with FetchContent

Fetch source dependencies at configure time so they build with your compiler/settings. Pin `GIT_TAG` to an exact tag or commit and add `GIT_SHALLOW TRUE`. `FetchContent_MakeAvailable` fetches and adds the subdir in one call — replaces the old `FetchContent_GetProperties`/`Populate` dance.

```cmake
include(FetchContent)
FetchContent_Declare(fmt
    GIT_REPOSITORY https://github.com/fmtlib/fmt.git
    GIT_TAG 10.1.1
    GIT_SHALLOW TRUE)
FetchContent_MakeAvailable(fmt)
target_link_libraries(app PRIVATE fmt::fmt)
```
