# DistConf C++ SDK

A modern C++11 header-only wrapper for `libdistconf`.

## Usage

```cpp
#include "DistConf.hpp"

using namespace distconf;

int main() {
    try {
        DistConfig cfg("standalone");
        
        // Register a listener (lambda support)
        cfg.OnLiveConfUpdate([](const std::string& json) {
            std::cout << "Update received: " << json << std::endl;
        });
        
        cfg.OnRegistryUpdate([](const std::string& json) {
            std::cout << "Registry changed: " << json << std::endl;
        });

        std::string name = cfg.Get("common", "name");
        cfg.Set("local", "status", "active");
        
        if (cfg.ValidateMandatoryServices()) {
            // ...
        }
    } catch (const std::exception& e) {
        // Handle error
    }
    return 0;
}
```

## Compilation

Link against `libdistconf.so`:
```bash
g++ main.cpp -L../libdistconf -ldistconf -o app
```
