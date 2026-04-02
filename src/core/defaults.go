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
			LogServer: &models.LogServerCapability{
				IP:   "${LS_IP:127.0.0.2}",
				Port: "${LS_PORT:9020}",
			},
			ConfigServer: &models.ConfigServerCapability{
				IP:      "${CF_IP:127.0.0.2}",
				Port:    "${CF_PORT:3306}",
				Refresh: "300",
			},
			NotifServer: &models.NotifServerCapability{
				IP:   "${NT_IP:127.0.0.2}",
				Port: "${NT_PORT:1026}",
			},
			TeleRemote: &models.TeleRemoteCapability{
				Token:  "${TR_TOKEN}",
				ChatID: "${TR_CHATID}",
				IP:     "${TR_IP:127.0.0.2}",
				Port:   "${TR_PORT:1863}",
			},
			WebInterface: &models.WebInterfaceCapability{
				IP:   "${WB_IP:127.0.0.2}",
				Port: "${WB_PORT:8080}",
			},
			Scheduler: &models.SchedulerCapability{
				IP:   "127.0.0.2",
				Port: "5001",
			},
			TimescaleDb: &models.TimescaleDbCapability{
				IP:       "${TS_IP:127.0.0.2}",
				Port:     "${TS_PORT:5432}",
				DBName:   "${TS_DBNAME:maindb}",
				User:     "${TS_USER}",
				Password: "${TS_PASSWORD}",
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
