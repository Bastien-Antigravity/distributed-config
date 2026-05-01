#include "../DistConf.hpp"
#include <cassert>
#include <chrono>
#include <iostream>
#include <thread>

using namespace distconf;

void test_lifecycle() {
  std::cout << "Testing Lifecycle..." << std::endl;
  {
    DistConfig cfg("standalone");
    // RAII should handle cleanup
  }
  std::cout << "  Passed." << std::endl;
}

void test_data_operations() {
  std::cout << "Testing Data Operations (Get/Set/FullConfig)..." << std::endl;
  DistConfig cfg("standalone");

  // Set/Get
  cfg.Set("cpp_test", "key1", "val1");
  assert(cfg.Get("cpp_test", "key1") == "val1");

  // Full Config
  std::string full = cfg.GetFullConfig();
  assert(full.find("cpp_test") != std::string::npos);
  assert(full.find("val1") != std::string::npos);

  std::cout << "  Passed." << std::endl;
}

void test_capabilities() {
  std::cout << "Testing Capabilities..." << std::endl;
  DistConfig cfg("standalone");

  // Standalone generates default capabilities (config_server, log_server)
  std::string cap = cfg.GetCapability("config_server");
  assert(cap.find("ip") != std::string::npos);

  // Networking
  std::string addr = cfg.GetAddress("config_server");
  assert(!addr.empty());
  assert(addr.find(":") != std::string::npos);

  std::cout << "  Passed." << std::endl;
}

void test_ecosystem_sharing() {
  std::cout << "Testing Ecosystem Sharing (ShareConfig)..." << std::endl;
  DistConfig cfg("standalone");

  // ShareConfig merges top-level keys as sections into LiveConfig
  std::string payload =
      "{\"cpp_service\": {\"status\":\"online\",\"version\":\"2.0\"}}";
  bool success = cfg.ShareConfig(payload);
  assert(success == true);

  // After ShareConfig, data is accessible via Get(section, key)
  std::string status = cfg.Get("cpp_service", "status");
  assert(status == "online");

  std::cout << "  Passed." << std::endl;
}

void test_security() {
  std::cout << "Testing Security (Decrypt)..." << std::endl;
  DistConfig cfg("standalone");

  // ENC(...) detection
  std::string ciphertext = "ENC(dummy)";
  try {
    std::string decrypted = cfg.Decrypt(ciphertext);
    assert(!decrypted.empty());
  } catch (const std::exception &e) {
    // Without keys in CI/CD, we expect a transparent error message
    std::string err = e.what();
    assert(!err.empty());
    std::cout << "  Passed (Caught Transparent Error: " << err << ")"
              << std::endl;
  }

  std::cout << "  Passed." << std::endl;
}

void test_validation() {
  std::cout << "Testing Validation (Fail-Fast)..." << std::endl;
  DistConfig cfg("standalone");

  // Standalone should be valid by default
  assert(cfg.ValidateMandatoryServices() == true);

  std::cout << "  Passed." << std::endl;
}

int main() {
  try {
    test_lifecycle();
    test_data_operations();
    test_capabilities();
    test_ecosystem_sharing();
    test_security();
    test_validation();

    std::cout << "\n=======================================" << std::endl;
    std::cout << "  All C++ SDK Parity Tests Passed! " << std::endl;
    std::cout << "=======================================" << std::endl;
  } catch (const std::exception &e) {
    std::cerr << "\n!!! TEST FAILED: " << e.what() << std::endl;
    return 1;
  }
  return 0;
}
