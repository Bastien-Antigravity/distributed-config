package core

// NewDefaultConfig returns a Config struct populated with default values.
// This replaces the External String Template to keep data and defaults "merged".
// -----------------------------------------------------------------------------

func NewDefaultConfig() *Config {
	return &Config{
		Common: CommonConfig{
			Name:  "common",
			Reset: false,
		},
		Capabilities: map[string]interface{}{
			"log_server": map[string]interface{}{
				"ip":   "${LS_IP:127.0.0.2}",
				"port": "${LS_PORT:9020}",
			},
			"config_server": map[string]interface{}{
				"ip":      "${CF_IP:127.0.0.2}",
				"port":    "${CF_PORT:3306}",
				"refresh": "300",
			},
			"notif_server": map[string]interface{}{
				"ip":   "${NT_IP:127.0.0.2}",
				"port": "${NT_PORT:1026}",
			},
			"tele_remote": map[string]interface{}{
				"token":   "${TR_TOKEN}",
				"chat_id": "${TR_CHATID}",
				"ip":      "${TR_IP:127.0.0.2}",
				"port":    "${TR_PORT:1863}",
			},
			"web_interface": map[string]interface{}{
				"ip":   "${WB_IP:127.0.0.2}",
				"port": "${WB_PORT:8080}",
			},
			"scheduler": map[string]interface{}{
				"ip":   "127.0.0.2",
				"port": "5001",
			},
			"timescale_db": map[string]interface{}{
				"ip":       "${TS_IP:127.0.0.2}",
				"port":     "${TS_PORT:5432}",
				"db_name":  "${TS_DBNAME:maindb}",
				"user":     "${TS_USER}",
				"password": "${TS_PASSWORD}",
				"ssl_cert": "false",
			},
			"file_system": map[string]interface{}{
				"temp_path": "./fs_temp",
				"data_path": "./fs_data",
			},
			"jupyter": map[string]interface{}{
				"ip":   "127.0.0.2",
				"port": "8888",
			},
		},
	}
}
