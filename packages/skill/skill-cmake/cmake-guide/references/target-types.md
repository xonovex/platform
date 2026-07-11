# target-types: Target Types

Pick the `add_library` kind by how it is consumed: no type = `BUILD_SHARED_LIBS` decides static/shared; `STATIC`/`SHARED` force it; `INTERFACE` for header-only (no compiled code); `OBJECT` to share compiled units across targets. Use `add_executable ... WIN32 MACOSX_BUNDLE` for GUI apps.

```cmake
add_library(mylib src/lib.cpp)                  # BUILD_SHARED_LIBS decides
add_library(static_lib STATIC src/static.cpp)
add_library(shared_lib SHARED src/shared.cpp)
add_library(header_only INTERFACE)
target_include_directories(header_only INTERFACE include)
add_library(objects OBJECT src/common.cpp)
target_link_libraries(mylib PRIVATE objects)
add_executable(gui WIN32 MACOSX_BUNDLE src/gui.cpp)
```
