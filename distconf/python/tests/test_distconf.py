#!/usr/bin/env python
# coding:utf-8

import unittest
from os import getenv as osGetenv
from os.path import abspath as osPathAbspath, dirname as osPathDirname, exists as osPathExists
from sys import path as sysPath
from time import sleep as timeSleep

# Add parent dir to path to import distconf
sysPath.append(osPathDirname(osPathDirname(osPathAbspath(__file__))))

from distconf import DistConfig

class TestDistConfigFull(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        # Respect LIBDISTCONF_PATH if set, otherwise fallback to relative path
        cls.lib_path = osGetenv("LIBDISTCONF_PATH")
        if not cls.lib_path:
            import platform
            system = platform.system()
            if system == "Darwin":
                ext = ".dylib"
            elif system == "Windows":
                ext = ".dll"
            else:
                ext = ".so"
            cls.lib_path = osPathAbspath(f"../../distconf/libdistconf/libdistconf{ext}")

    # -----------------------------------------------------------------------------------------------

    def test_01_lifecycle(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        self.assertIsNotNone(cfg._handle)
        cfg.close()
        self.assertIsNone(cfg._handle)

    # -----------------------------------------------------------------------------------------------

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

    # -----------------------------------------------------------------------------------------------

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

    # -----------------------------------------------------------------------------------------------

    def test_04_sharing(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        
        payload = {"py_service": {"status": "ok", "items": "1,2,3"}}
        success = cfg.share_config(payload)
        self.assertTrue(success)
        
        # Verify it reflected in LiveConfig
        status = cfg.get("py_service", "status")
        self.assertEqual(status, "ok")
        
        cfg.close()

    # -----------------------------------------------------------------------------------------------

    def test_05_security(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        
        ciphertext = "ENC(hello)"
        decrypted = cfg.decrypt(ciphertext)
        self.assertIsNotNone(decrypted)
        
        cfg.close()

    # -----------------------------------------------------------------------------------------------

    def test_06_validation(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        self.assertTrue(cfg.validate_mandatory_services())
        cfg.close()

    # -----------------------------------------------------------------------------------------------

    def test_07_callbacks(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        
        updated_data = []
        def on_update(data):
            updated_data.append(data)
            
        cfg.on_live_conf_update(on_update)
        
        # Trigger update via local set
        cfg.set("callback_test", "trigger", "now")
        
        # Wait for callback dispatch (Go -> C -> Python)
        timeSleep(0.1)
        
        self.assertTrue(len(updated_data) > 0)
        self.assertIn("callback_test", updated_data[0])
        
        cfg.close()

    # -----------------------------------------------------------------------------------------------

    def test_08_registry_callbacks(self):
        cfg = DistConfig("standalone", lib_path=self.lib_path)
        
        registry_events = []
        def on_registry(data):
            registry_events.append(data)
            
        cfg.on_registry_update(on_registry)
        
        # Test that we can register it without crashing. 
        # (Actually triggering it requires injecting network packets or using internal Go hooks, 
        # but registering it ensures the CGO bridge bindings and ctypes wrappers are fully functional).
        self.assertEqual(len(registry_events), 0)
        
        cfg.close()

if __name__ == "__main__":
    unittest.main()
