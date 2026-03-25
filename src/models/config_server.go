package models

// Config Server Capability
// -----------------------------------------------------------------------------

type ConfigServerCapability struct {
	IP      string `yaml:"ip"`
	Port    string `yaml:"port"`
	Refresh string `yaml:"refresh"`
}
