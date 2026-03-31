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
	Logger       *models.LoggerCapability       `yaml:"logger" json:"logger,omitempty"`
	ConfigServer *models.ConfigServerCapability `yaml:"config_server" json:"config_server,omitempty"`
	Notification *models.NotificationCapability `yaml:"notification" json:"notification,omitempty"`
	Telebot      *models.TelebotCapability      `yaml:"telebot" json:"telebot,omitempty"`
	Scheduler    *models.SchedulerCapability    `yaml:"scheduler" json:"scheduler,omitempty"`
	Monitoring   *models.MonitoringCapability   `yaml:"monitoring" json:"monitoring,omitempty"`
	Database     *models.DatabaseCapability     `yaml:"database" json:"database,omitempty"`
	FileSystem   *models.FileSystemCapability   `yaml:"file_system" json:"file_system,omitempty"`
	Jupyter      *models.JupyterCapability      `yaml:"jupyter" json:"jupyter,omitempty"`
}
