package models

// Database Capability
// -----------------------------------------------------------------------------

type DatabaseCapability struct {
	IP       string `yaml:"ip"`
	Port     string `yaml:"port"`
	DBName   string `yaml:"db_name"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	SSLCert  string `yaml:"ssl_cert"`
}
