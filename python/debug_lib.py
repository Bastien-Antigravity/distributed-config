import os
import sys
from distributed_config import DistributedConfig

def main():
    print("Initializing DistributedConfig...")
    try:
        with DistributedConfig("test") as dc:
            print(f"Handle: {dc.handle}")
            cfg = dc.get_config()
            print("Config retrieved successfully!")
            print(f"Common Name: {cfg['common']['name']}")
    except Exception as e:
        print(f"Error: {e}")

if __name__ == "__main__":
    # Ensure the python path includes the current directory
    sys.path.insert(0, os.path.abspath(os.path.dirname(__file__)))
    main()
