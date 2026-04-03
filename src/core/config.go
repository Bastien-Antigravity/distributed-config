package core

import "github.com/Bastien-Antigravity/distributed-config/src/models"

// Common Config
// -----------------------------------------------------------------------------

type CommonConfig struct {
	Name  string `yaml:"name" json:"name"`
	Reset bool   `yaml:"reset" json:"reset"`
}

// Config Data Struct (Pure Data)
// -----------------------------------------------------------------------------

type Config struct {
	Common       CommonConfig `yaml:"common" json:"common"`
	Capabilities Capabilities `yaml:"capabilities" json:"capabilities"`

	// Internal state
	COMMON_FILE_PATH string
	MemConfigPath    string

	// Data storage for MemConfig
	MemConfig map[string]map[string]string `yaml:"-"`
}

// Capabilities Container
// -----------------------------------------------------------------------------

type Capabilities struct {
	LogServer    *models.LogServerCapability    `yaml:"log_server" json:"log_server,omitempty"`
	ConfigServer *models.ConfigServerCapability `yaml:"config_server" json:"config_server,omitempty"`
	NotifServer  *models.NotifServerCapability  `yaml:"notif_server" json:"notif_server,omitempty"`
	TeleRemote   *models.TeleRemoteCapability   `yaml:"tele_remote" json:"tele_remote,omitempty"`
	Scheduler    *models.SchedulerCapability    `yaml:"scheduler" json:"scheduler,omitempty"`
	WebInterface *models.WebInterfaceCapability `yaml:"web_interface" json:"web_interface,omitempty"`
	TimescaleDb  *models.TimescaleDbCapability  `yaml:"timescale_db" json:"timescale_db,omitempty"`
	FileSystem   *models.FileSystemCapability   `yaml:"file_system" json:"file_system,omitempty"`
	Jupyter      *models.JupyterCapability      `yaml:"jupyter" json:"jupyter,omitempty"`
}

// Get returns a value from a specified section and key.
// Returns an empty string if not found.
func (c *Config) Get(section, key string) string {
	if s, ok := c.MemConfig[section]; ok {
		if val, ok := s[key]; ok {
			return val
		}
	}
	return ""
}

// Set sets a value for a specified section and key.
// Initializes MemConfig and section maps if they are nil.
func (c *Config) Set(section, key, value string) {
	if c.MemConfig == nil {
		c.MemConfig = make(map[string]map[string]string)
	}
	if _, ok := c.MemConfig[section]; !ok {
		c.MemConfig[section] = make(map[string]string)
	}
	c.MemConfig[section][key] = value
}
