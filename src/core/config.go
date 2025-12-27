package core

import "github.com/Bastien-Antigravity/distributed-config/src/models"

// Common Config
// -----------------------------------------------------------------------------

type CommonConfig struct {
	Name  string `yaml:"name"`
	Reset bool   `yaml:"reset"`
}

// Config Data Struct (Pure Data)
// -----------------------------------------------------------------------------

type Config struct {
	Common       CommonConfig `yaml:"common"`
	Capabilities Capabilities `yaml:"capabilities"`

	// Internal state
	COMMON_FILE_PATH string
	MemConfigPath    string

	// Data storage for MemConfig
	MemConfig map[string]map[string]string `yaml:"-"`
}

// Capabilities Container
// -----------------------------------------------------------------------------

type Capabilities struct {
	Logger       *models.LoggerCapability       `yaml:"logger"`
	ConfigServer *models.ConfigServerCapability `yaml:"config_server"`
	Notification *models.NotificationCapability `yaml:"notification"`
	Telebot      *models.TelebotCapability      `yaml:"telebot"`
	Scheduler    *models.SchedulerCapability    `yaml:"scheduler"`
	Monitoring   *models.MonitoringCapability   `yaml:"monitoring"`
	Database     *models.DatabaseCapability     `yaml:"database"`
	FileSystem   *models.FileSystemCapability   `yaml:"file_system"`
	Jupyter      *models.JupyterCapability      `yaml:"jupyter"`
}
