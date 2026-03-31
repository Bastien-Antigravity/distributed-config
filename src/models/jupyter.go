package models

// Jupyter Capability
// -----------------------------------------------------------------------------

type JupyterCapability struct {
	IP   string `yaml:"ip" json:"ip"`
	Port string `yaml:"port" json:"port"`
}
