package models

// Config Server Capability
// -----------------------------------------------------------------------------

type ConfigServerCapability struct {
	IP      string `yaml:"ip" json:"ip"`
	Port    string `yaml:"port" json:"port"`
	Refresh string `yaml:"refresh" json:"refresh"`
}
