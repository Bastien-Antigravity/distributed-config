package models

// File System Capability
// -----------------------------------------------------------------------------

type FileSystemCapability struct {
	TempPath string `yaml:"temp_path" json:"temp_path"`
	DataPath string `yaml:"data_path" json:"data_path"`
}
