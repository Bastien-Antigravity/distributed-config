#ifndef DISTCONF_HPP
#define DISTCONF_HPP

#include <string>
#include <vector>
#include <functional>
#include <stdexcept>
#include <memory>

// Include the generated C header
#include "../libdistconf/libdistconf.h"
#include <map>
#include <mutex>

namespace distconf {

/**
 * DistConfig is a C++ wrapper around the libdistconf CGO bridge.
 * It provides a clean, object-oriented interface for configuration management.
 */
class DistConfig {
public:
    explicit DistConfig(const std::string& profile) {
        handle_ = DistConf_New(const_cast<char*>(profile.c_str()));
        if (handle_ == 0) {
            throw std::runtime_error("Failed to initialize DistConf with profile: " + profile);
        }
    }

    ~DistConfig() {
        if (handle_ != 0) {
            DistConf_Close(handle_);
        }
    }

    // Disable copy
    DistConfig(const DistConfig&) = delete;
    DistConfig& operator=(const DistConfig&) = delete;

    // Get a configuration value
    std::string Get(const std::string& section, const std::string& key) const {
        char* val = DistConf_Get(handle_, 
                                 const_cast<char*>(section.c_str()), 
                                 const_cast<char*>(key.c_str()));
        if (!val) return "";
        std::string result(val);
        DistConf_FreeString(val);
        return result;
    }

    // Set a configuration value
    void Set(const std::string& section, const std::string& key, const std::string& value) {
        DistConf_Set(handle_, 
                     const_cast<char*>(section.c_str()), 
                     const_cast<char*>(key.c_str()), 
                     const_cast<char*>(value.c_str()));
    }

    // Synchronize with the Config Server
    bool Sync() {
        return DistConf_Sync(handle_) != 0;
    }

    // Broadcast state to the ecosystem
    bool ShareObject(const std::string& section, const std::string& json_data) {
        return DistConf_ShareObject(handle_, 
                                  const_cast<char*>(section.c_str()), 
                                  const_cast<char*>(json_data.c_str())) != 0;
    }

    // Validate mandatory services
    bool ValidateMandatoryServices() {
        return DistConf_ValidateMandatoryServices(handle_) != 0;
    }

    // Get an address (host:port) for a capability
    std::string GetAddress(const std::string& capability) const {
        char* val = DistConf_GetAddress(handle_, const_cast<char*>(capability.c_str()));
        if (!val) return "";
        std::string result(val);
        DistConf_FreeString(val);
        return result;
    }

    // Get a gRPC address for a capability
    std::string GetGRPCAddress(const std::string& capability) const {
        char* val = DistConf_GetGRPCAddress(handle_, const_cast<char*>(capability.c_str()));
        if (!val) return "";
        std::string result(val);
        DistConf_FreeString(val);
        return result;
    }

    // Get full configuration as JSON
    std::string GetFullConfig() const {
        char* val = DistConf_GetFullConfig(handle_);
        if (!val) return "{}";
        std::string result(val);
        DistConf_FreeString(val);
        return result;
    }

    // Get a capability configuration as JSON
    std::string GetCapability(const std::string& capability) const {
        char* val = DistConf_GetCapability(handle_, const_cast<char*>(capability.c_str()));
        if (!val) return "{}";
        std::string result(val);
        DistConf_FreeString(val);
        return result;
    }

    // Decrypt a secret
    std::string Decrypt(const std::string& ciphertext) const {
        char* val = DistConf_Decrypt(handle_, const_cast<char*>(ciphertext.c_str()));
        if (!val) return ciphertext;
        std::string result(val);
        DistConf_FreeString(val);
        return result;
    }

    // Register a live update listener
    void OnLiveConfUpdate(std::function<void(const std::string&)> callback) {
        callback_ = callback;
        
        // Register this instance in the global registry
        std::lock_guard<std::mutex> lock(registry_mutex_);
        registry_[handle_] = this;

        // Register the static bridge with libdistconf
        DistConf_OnLiveConfUpdate(handle_, StaticCallbackBridge);
    }

private:
    static void StaticCallbackBridge(GoUintptr handle, const char* json_data) {
        std::lock_guard<std::mutex> lock(registry_mutex_);
        auto it = registry_.find(handle);
        if (it != registry_.end() && it->second->callback_) {
            it->second->callback_(std::string(json_data));
        }
    }

    uintptr_t handle_;
    std::function<void(const std::string&)> callback_;

    // Static registry to route C callbacks to the correct DistConfig instance
    static std::map<uintptr_t, DistConfig*> registry_;
    static std::mutex registry_mutex_;
};

// Initialize static members
inline std::map<uintptr_t, DistConfig*> DistConfig::registry_;
inline std::mutex DistConfig::registry_mutex_;

} // namespace distconf

#endif // DISTCONF_HPP
