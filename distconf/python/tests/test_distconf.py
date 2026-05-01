import unittest
import os
import sys
import json
import time

# Add parent dir to path to import distconf
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from distconf import DistConfig

class TestDistConfigFull(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        # Path to the library built by the Makefile
        cls.lib_path = os.path.abspath("../../distconf/libdistconf/libdistconf.so")
        if not os.path.exists(cls.lib_path):
            # Try dylib for macOS
            cls.lib_path = cls.lib_path.replace(".so", ".dylib")
            
    def test_01_lifecycle(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        self.assertIsNotNone(cfg._handle)
        cfg.close()
        self.assertIsNone(cfg._handle)

    def test_02_data_access(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        
        # Basic Set/Get
        cfg.set("py_test", "key1", "value1")
        self.assertEqual(cfg.get("py_test", "key1"), "value1")
        
        # Get Full Config
        full = cfg.get_full_config()
        self.assertIn("py_test", full.get("live", {}))
        self.assertEqual(full["live"]["py_test"]["key1"], "value1")
        
        cfg.close()

    def test_03_capabilities(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        
        # Resolve address
        addr = cfg.get_address("config_server")
        self.assertTrue(":" in addr)
        
        # Get specific capability
        cap = cfg.get_capability("config_server")
        self.assertIn("ip", cap)
        self.assertIn("port", cap)
        
        cfg.close()

    def test_04_sharing(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        
        payload = {"status": "ok", "items": [1, 2, 3]}
        success = cfg.share_object("py_service", payload)
        self.assertTrue(success)
        
        # Verify it reflected in LiveConfig
        shared = cfg.get("py_service", "shared_data")
        shared_parsed = json.loads(shared)
        self.assertEqual(shared_parsed["status"], "ok")
        
        cfg.close()

    def test_05_security(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        
        ciphertext = "ENC(hello)"
        decrypted = cfg.decrypt(ciphertext)
        self.assertIsNotNone(decrypted)
        
        cfg.close()

    def test_06_validation(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        self.assertTrue(cfg.validate_mandatory_services())
        cfg.close()

    def test_07_callbacks(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        
        updated_data = []
        def on_update(data):
            updated_data.append(data)
            
        cfg.on_live_conf_update(on_update)
        
        # Trigger update via local set
        cfg.set("callback_test", "trigger", "now")
        
        # Wait for callback dispatch (Go -> C -> Python)
        time.sleep(0.1)
        
        self.assertTrue(len(updated_data) > 0)
        self.assertIn("callback_test", updated_data[0])
        
        cfg.close()

if __name__ == "__main__":
    unittest.main()
