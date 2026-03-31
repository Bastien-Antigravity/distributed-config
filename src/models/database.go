package models

// Database Capability
// -----------------------------------------------------------------------------

type DatabaseCapability struct {
	IP       string `yaml:"ip" json:"ip"`
	Port     string `yaml:"port" json:"port"`
	DBName   string `yaml:"db_name" json:"db_name"`
	User     string `yaml:"user" json:"user"`
	Password string `yaml:"password" json:"password"`
	SSLCert  string `yaml:"ssl_cert" json:"ssl_cert"`
}
