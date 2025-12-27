package models

// Database Capability
// -----------------------------------------------------------------------------

type DatabaseCapability struct {
	IP       string `yaml:"ip"`
	Port     string `yaml:"port"`
	Endpoint string `yaml:"endpoint"`
	SSLCert  string `yaml:"ssl_cert"`
	Backend  string `yaml:"backend"`
}
