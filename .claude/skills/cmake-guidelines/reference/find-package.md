# find-package: Dependency Management with find_package

**Guideline:** Use find_package to locate and link system-installed dependencies.

**Rationale:** System-installed packages are pre-built and optimized for the platform, reducing build times and disk usage when available.

**Example:**

```cmake
# Find system-installed package
find_package(Boost 1.80 REQUIRED COMPONENTS system filesystem)

add_executable(app src/main.cpp)
target_link_libraries(app PRIVATE Boost::system Boost::filesystem)

# Optional dependency
find_package(OpenSSL)
if(OpenSSL_FOUND)
    target_link_libraries(app PRIVATE OpenSSL::SSL OpenSSL::Crypto)
    target_compile_definitions(app PRIVATE HAVE_OPENSSL=1)
endif()
```

**Techniques:**
- REQUIRED: Fail build if package not found, use for mandatory dependencies
- COMPONENTS: Specify individual package parts to reduce linking
- Namespaced targets: Use Boost::system, OpenSSL::SSL for safe target linking
- \*_FOUND variable: Check if optional dependency is available before using
- Conditional linking: Add target_compile_definitions when dependencies found
