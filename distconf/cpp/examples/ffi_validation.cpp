#include <iostream>
#include <string>
#include <dlfcn.h>
#include <assert.h>

typedef size_t (*DistConf_New_t)(const char* profile);
typedef char* (*DistConf_Get_t)(size_t handle, const char* section, const char* key);
typedef int (*DistConf_Set_t)(size_t handle, const char* section, const char* key, const char* value);
typedef void (*DistConf_Close_t)(size_t handle);
typedef void (*DistConf_FreeString_t)(char* ptr);

int main() {
    std::cout << "--- FFI Validation: C++ ---" << std::endl;

    const char* lib_path = "../../distconf/libdistconf/libdistconf.dylib";
#ifdef __linux__
    lib_path = "../../distconf/libdistconf/libdistconf.so";
#endif

    void* handle_lib = dlopen(lib_path, RTLD_LAZY);
    if (!handle_lib) {
        std::cerr << "Skipping: Library not found at " << lib_path << " (" << dlerror() << ")" << std::endl;
        return 0;
    }

    auto dist_conf_new = (DistConf_New_t)dlsym(handle_lib, "DistConf_New");
    auto dist_conf_get = (DistConf_Get_t)dlsym(handle_lib, "DistConf_Get");
    auto dist_conf_set = (DistConf_Set_t)dlsym(handle_lib, "DistConf_Set");
    auto dist_conf_close = (DistConf_Close_t)dlsym(handle_lib, "DistConf_Close");
    auto dist_conf_free_string = (DistConf_FreeString_t)dlsym(handle_lib, "DistConf_FreeString");

    // 1. Initialize
    size_t handle = dist_conf_new("standalone");
    if (handle == 0) {
        std::cerr << "FAIL: Failed to create handle" << std::endl;
        return 1;
    }
    std::cout << "OK: Created handle " << handle << std::endl;

    // 2. Set/Get
    dist_conf_set(handle, "section1", "key1", "value-cpp");
    char* val = dist_conf_get(handle, "section1", "key1");
    if (val && std::string(val) == "value-cpp") {
        std::cout << "OK: Get returned expected value: " << val << std::endl;
    } else {
        std::cerr << "FAIL: Get returned " << (val ? val : "NULL") << std::endl;
        return 1;
    }

    // 3. Cleanup
    dist_conf_free_string(val);
    dist_conf_close(handle);
    dlclose(handle_lib);
    std::cout << "OK: Closed handle and freed strings" << std::endl;

    std::cout << "--- FFI Validation Success ---" << std::endl;
    return 0;
}
