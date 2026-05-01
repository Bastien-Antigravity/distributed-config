#include "../DistConf.hpp"
#include <iostream>
#include <string>

using namespace distconf;

int main() {
    try {
        std::cout << "--- DistConf C++ SDK Demo ---" << std::endl;

        // 1. Initialize configuration for "standalone" environment
        distconf::DistConfig cfg("standalone");

        // 1.5 Register a callback for live updates
        cfg.OnLiveConfUpdate([](const std::string& jsonData) {
            std::cout << "[CALLBACK] Live Config Updated! New State: " << jsonData << std::endl;
        });

        // 2. Get static values
        std::string appName = cfg.Get("common", "name");
        std::cout << "App Name: " << appName << std::endl;

        // 3. Set local state
        cfg.Set("cpp_demo", "status", "active");
        std::cout << "Local Status: " << cfg.Get("cpp_demo", "status") << std::endl;

        // 4. Share object with ecosystem
        std::string healthJson = "{\"status\": \"ok\", \"threads\": 8}";
        if (cfg.ShareObject("health", healthJson)) {
            std::cout << "Successfully shared health JSON." << std::endl;
        }

        // 5. Validate environment
        if (cfg.ValidateMandatoryServices()) {
            std::cout << "Mandatory services validation passed." << std::endl;
        }

        // 6. Decrypt secrets
        std::string encrypted = "ENC(SGVsbG8=)";
        std::cout << "Decrypted: " << cfg.Decrypt(encrypted) << std::endl;

    } catch (const std::exception& e) {
        std::cerr << "CRITICAL ERROR: " << e.what() << std::endl;
        return 1;
    }

    std::cout << "--- Demo Completed ---" << std::endl;
    return 0;
}
