---
name: cmake-guidelines
description: >-
  Trigger on `CMakeLists.txt` files. Use when writing CMake 3.20+ build systems for C/C++ projects. Apply for target-based builds, dependency management, testing setup, visibility specifiers. Keywords: CMake, add_library, add_executable, target_link_libraries, PUBLIC/PRIVATE/INTERFACE, find_package, FetchContent, CTest.
---

# CMake Coding Guidelines

## Requirements

- CMake ≥ 3.20; modern target-based usage.

## Essentials

- **Target-based builds** - Use targets, no global include/link dirs, see [reference/target-types.md](reference/target-types.md), [reference/compile-options.md](reference/compile-options.md)
- **Visibility specifiers** - Use PUBLIC/PRIVATE/INTERFACE correctly, see [reference/visibility-specifiers.md](reference/visibility-specifiers.md)
- **Dependencies** - Declare explicitly with FetchContent/find_package, see [reference/find-package.md](reference/find-package.md), [reference/fetchcontent.md](reference/fetchcontent.md)
- **Testing** - Enable testing with CTest, see [reference/testing.md](reference/testing.md)
- **Project structure** - Organize multi-directory projects, see [reference/project-structure.md](reference/project-structure.md)

## Progressive disclosure

- Read [reference/target-types.md](reference/target-types.md) - When choosing between library types or executables
- Read [reference/visibility-specifiers.md](reference/visibility-specifiers.md) - When deciding PUBLIC vs PRIVATE vs INTERFACE
- Read [reference/compile-options.md](reference/compile-options.md) - When adding compiler flags to targets
- Read [reference/find-package.md](reference/find-package.md) - When integrating external dependencies
- Read [reference/fetchcontent.md](reference/fetchcontent.md) - When vendoring dependencies from git/archives
- Read [reference/testing.md](reference/testing.md) - When setting up CTest or test targets
- Read [reference/project-structure.md](reference/project-structure.md) - When organizing multi-directory CMake projects
- Read [reference/generator-expressions.md](reference/generator-expressions.md) - When using conditional build configuration
- Read [reference/installation.md](reference/installation.md) - When creating install targets or package exports
