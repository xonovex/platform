---
name: cmake-guide
description: "Use when editing CMake build files for C/C++ projects on CMake 3.20+. Triggers on `CMakeLists.txt`, `*.cmake` files and on prompts about adding targets, linking libraries, dependency fetching, test setup, install rules, or PUBLIC/PRIVATE/INTERFACE visibility, even when the user doesn't say 'CMake'. Apply for FetchContent, find_package, target_link_libraries, CTest. Skip Make, Bazel, Meson, and other build systems."
---

# CMake Coding Guidelines

## Requirements

- CMake ≥ 3.20; modern target-based usage.

## Essentials

- **Target-based builds** - Use targets, no global include/link dirs, see [references/target-types.md](references/target-types.md), [references/compile-options.md](references/compile-options.md)
- **Visibility specifiers** - Use PUBLIC/PRIVATE/INTERFACE correctly, see [references/visibility-specifiers.md](references/visibility-specifiers.md)
- **Dependencies** - Declare explicitly with FetchContent/find_package, see [references/find-package.md](references/find-package.md), [references/fetchcontent.md](references/fetchcontent.md)
- **Testing** - Enable testing with CTest, see [references/testing.md](references/testing.md)
- **Project structure** - Organize multi-directory projects, see [references/project-structure.md](references/project-structure.md)

## Gotchas

- `target_link_libraries` scope matters: PRIVATE = consumers don't see it, INTERFACE = no compile, PUBLIC = both — wrong scope leaks transitive deps
- `find_package` may use PATHS or HINTS but ignores both if a config file is on a system path; use `<Pkg>_ROOT` env var to force-locate
- Generator expressions (`$<CONFIG:Debug>`) only evaluate at build time — debugging by `message()` won't show their final values
- `CMAKE_INSTALL_PREFIX` is captured at configure time; changing it after first config requires a clean reconfigure

## Progressive disclosure

- Read [references/target-types.md](references/target-types.md) - When choosing between library types or executables
- Read [references/visibility-specifiers.md](references/visibility-specifiers.md) - When deciding PUBLIC vs PRIVATE vs INTERFACE
- Read [references/compile-options.md](references/compile-options.md) - When adding compiler flags to targets
- Read [references/find-package.md](references/find-package.md) - When integrating external dependencies
- Read [references/fetchcontent.md](references/fetchcontent.md) - When vendoring dependencies from git/archives
- Read [references/testing.md](references/testing.md) - When setting up CTest or test targets
- Read [references/project-structure.md](references/project-structure.md) - When organizing multi-directory CMake projects
- Read [references/generator-expressions.md](references/generator-expressions.md) - When using conditional build configuration
- Read [references/installation.md](references/installation.md) - When creating install targets or package exports
