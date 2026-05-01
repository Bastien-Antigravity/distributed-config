import os
import sys

# Add parent directory to path so we can import distconf
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from distconf import DistConfig

def main():
    # 1. Initialize with a profile
    # For this example, we use 'standalone' which doesn't require a network.
    print("--- Initializing DistConf ---")
    
    # Path to the library (relative to this example)
    lib_path = os.path.abspath("../libdistconf/libdistconf.so")
    if not os.path.exists(lib_path):
        lib_path = lib_path.replace(".so", ".dylib")

    cfg = DistConfig("standalone", lib_path=lib_path)
    
    # 2. Get static configuration
    app_name = cfg.get("common", "name")
    print(f"Application Name: {app_name}")
    
    # 3. Set local configuration (triggers local callbacks)
    print("\n--- Setting Local Config ---")
    cfg.set("local_state", "status", "ready")
    status = cfg.get("local_state", "status")
    print(f"Status: {status}")
    
    # 4. Broadcast state to the ecosystem (ShareObject)
    print("\n--- Sharing Object ---")
    payload = {
        "service": "python-demo",
        "health": "excellent",
        "load": 0.42
    }
    if cfg.share_object("health_monitor", payload):
        print("Successfully shared health data.")
        
    # 5. Validate environment
    print("\n--- Validating Environment ---")
    if cfg.validate_mandatory_services():
        print("All mandatory services are correctly configured.")
    else:
        print("Validation failed: Some mandatory services are missing.")
        
    # 6. Decrypt a secret (if keys are present)
    secret_token = "ENC(SGVsbG8gV29ybGQh)" # Example dummy token
    decrypted = cfg.decrypt(secret_token)
    print(f"\nSecret Token: {decrypted}")
    
    # 7. Cleanup
    cfg.close()
    print("\n--- Done ---")

if __name__ == "__main__":
    main()
