# find-package: Dependency Management with find_package

Locate system deps with `find_package`; link namespaced imported targets (`Boost::system`), never bare names. `REQUIRED` fails hard if missing; without it, gate use on `<Pkg>_FOUND`. Narrow with `COMPONENTS`.

```cmake
find_package(Boost 1.80 REQUIRED COMPONENTS system filesystem)
target_link_libraries(app PRIVATE Boost::system Boost::filesystem)

find_package(OpenSSL)
if(OpenSSL_FOUND)
    target_link_libraries(app PRIVATE OpenSSL::SSL OpenSSL::Crypto)
    target_compile_definitions(app PRIVATE HAVE_OPENSSL=1)
endif()
```
