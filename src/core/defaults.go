package core

// NewDefaultConfig returns a Config struct populated with default values.
// This replaces the External String Template to keep data and defaults "merged".
// -----------------------------------------------------------------------------

func NewDefaultConfig() *Config {
	return &Config{
		Common: CommonConfig{
			Name:        "common",
			Reset:       false,
			PublicIP:    "127.0.0.1",
			RetryBaseMS: "100",
			RetryMaxSec: "5",
		},
		Capabilities: map[string]interface{}{
			"config_server": map[string]interface{}{
				"ip":        "${CF_IP:127.0.0.1}",
				"port":      "${CF_PORT:3306}",
				"grpc_ip":   "${CF_GRPC_IP:127.0.0.1}",
				"grpc_port": "${CF_GRPC_PORT:3307}",
				"refresh":   "300",
			},
			"log_server": map[string]interface{}{
				"ip":        "${LS_IP:127.0.0.1}",
				"port":      "${LS_PORT:9020}",
				"grpc_ip":   "${LS_GRPC_IP:127.0.0.1}",
				"grpc_port": "${LS_GRPC_PORT:9021}",
			},
			"notif_server": map[string]interface{}{
				"ip":        "${NT_IP:127.0.0.1}",
				"port":      "${NT_PORT:1026}",
				"grpc_ip":   "${NT_GRPC_IP:127.0.0.1}",
				"grpc_port": "${NT_GRPC_PORT:1027}",
			},
			"tele_remote": map[string]interface{}{
				"ip":      "${TR_IP:127.0.0.1}",
				"port":    "${TR_PORT:1863}",
				"token":   "${TR_TOKEN}",
				"chat_id": "${TR_CHATID}",
				"url":     "${TR_URL:https://api.telegram.org}",
			},
			"nats": map[string]interface{}{
				"servers":        []string{"nats://${NATS_IP:127.0.0.1}:4222"},
				"client_id":      "${APP_NAME:generic_client}",
				"subject_prefix": "antigravity",
			},
			"timescale_db": map[string]interface{}{
				"ip":       "${TS_IP:127.0.0.1}",
				"port":     "${TS_PORT:5432}",
				"db_name":  "${TS_DBNAME:maindb}",
				"user":     "${TS_USER:dbuser}",
				"password": "${TS_PASSWORD:dbuser}",
				"ssl_cert": "false",
			},
			"ontime_scheduler": map[string]interface{}{
				"ip":        "${SC_IP:127.0.0.1}",
				"port":      "${SC_PORT:8080}",
				"grpc_ip":   "${SC_GRPC_IP:127.0.0.1}",
				"grpc_port": "${SC_GRPC_PORT:8081}",
			},
			"web_interface": map[string]interface{}{
				"ip":   "${WB_IP:127.0.0.1}",
				"port": "${WB_PORT:8080}",
			},
			"file_system": map[string]interface{}{
				"temp_path": "./fs_temp",
				"data_path": "./fs_data",
			},
			"jupyter": map[string]interface{}{
				"ip":   "127.0.0.1",
				"port": "8888",
			},
		},
	}
}
