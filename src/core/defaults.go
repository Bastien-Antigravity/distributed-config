package core

import "github.com/Bastien-Antigravity/distributed-config/src/models"

// NewDefaultConfig returns a Config struct populated with default values.
// This replaces the External String Template to keep data and defaults "merged".
// -----------------------------------------------------------------------------

func NewDefaultConfig() *Config {
	return &Config{
		Common: CommonConfig{
			Name:  "common",
			Reset: false,
		},
		Capabilities: Capabilities{
			Logger: &models.LoggerCapability{
				IP:   "127.0.0.2",
				Port: "9020",
			},
			ConfigServer: &models.ConfigServerCapability{
				IP:      "127.0.0.2",
				Port:    "1026",
				Refresh: "300",
			},
			Notification: &models.NotificationCapability{
				IP:   "127.0.0.2",
				Port: "10080",
			},
			Telebot: &models.TelebotCapability{
				Token:  "${TELEBOT_TOKEN}",
				ChatID: "${TELEBOT_CHAT_ID}",
				IP:     "127.0.0.2",
				Port:   "31337",
			},
			Scheduler: &models.SchedulerCapability{
				IP:   "127.0.0.2",
				Port: "5001",
			},
			Monitoring: &models.MonitoringCapability{
				IP:   "127.0.0.2",
				Port: "5000",
			},
			Database: &models.DatabaseCapability{
				IP:       "127.0.0.2",
				Port:     "5432",
				DBName:   "maindb",
				User:     "${DB_USER}",
				Password: "${DB_PASSWORD}",
				SSLCert:  "false",
			},
			FileSystem: &models.FileSystemCapability{
				TempPath: "./fs_temp",
				DataPath: "./fs_data",
			},
			Jupyter: &models.JupyterCapability{
				IP:   "127.0.0.2",
				Port: "8888",
			},
		},
	}
}
