import pytest
import os
import sys

# Add the python directory to sys.path to allow importing the package
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))

from distributed_config import DistributedConfig

def test_initialization():
    """Test that the library can be initialized with the 'test' profile."""
    try:
        with DistributedConfig("test") as dc:
            assert dc.profile == "test"
            assert dc.handle > 0
    except FileNotFoundError:
        pytest.skip("Shared library not found. Run 'make build' first.")

def test_get_config():
    """Test that we can retrieve configuration as a dictionary."""
    try:
        with DistributedConfig("test") as dc:
            cfg = dc.get_config()
            
            # Check basic structure
            assert "common" in cfg
            assert "capabilities" in cfg
            
            # Check some default values from core/defaults.go
            assert cfg["common"]["name"] == "common"
            assert cfg["capabilities"]["logger"]["ip"] == "127.0.0.2"
            assert cfg["capabilities"]["database"]["db_name"] == "maindb"
            
    except FileNotFoundError:
        pytest.skip("Shared library not found. Run 'make build' first.")

def test_multiple_instances():
    """Test that multiple instances can coexist."""
    try:
        dc1 = DistributedConfig("test")
        dc2 = DistributedConfig("standalone")
        
        assert dc1.handle != dc2.handle
        
        dc1.close()
        dc2.close()
    except FileNotFoundError:
        pytest.skip("Shared library not found. Run 'make build' first.")
