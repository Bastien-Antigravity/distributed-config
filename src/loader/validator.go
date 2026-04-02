package loader

import (
	"fmt"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
)

// ValidateCommonConfig checks if the minimal required configuration is present.
// -----------------------------------------------------------------------------

func ValidateCommonConfig(config *core.Config) error {
	if config.Common.Name == "" {
		return fmt.Errorf("missing required common config: NAME")
	}

	// Assuming Config Server details are required for Distributed Mode
	if config.Capabilities.ConfigServer == nil || config.Capabilities.ConfigServer.IP == "" || config.Capabilities.ConfigServer.Port == "" {
		return fmt.Errorf("missing required config server details (CF_IP, CF_PORT) for distributed mode")
	}

	return nil
}

// CheckTestIPs ensures all defined IPs are 127.0.0.2
// -----------------------------------------------------------------------------

func CheckTestIPs(config *core.Config) error {
	// Helper to check individual IP
	check := func(name, ip string) error {
		if ip != "" && ip != "127.0.0.2" {
			return fmt.Errorf("test integrity failure: %s IP must be 127.0.0.2, got %s", name, ip)
		}
		return nil
	}

	caps := config.Capabilities
	if caps.LogServer != nil {
		if err := check("log_server", caps.LogServer.IP); err != nil {
			return err
		}
	}
	if caps.ConfigServer != nil {
		if err := check("config_server", caps.ConfigServer.IP); err != nil {
			return err
		}
	}
	if caps.NotifServer != nil {
		if err := check("notif_server", caps.NotifServer.IP); err != nil {
			return err
		}
	}
	if caps.TeleRemote != nil {
		if err := check("tele_remote", caps.TeleRemote.IP); err != nil {
			return err
		}
	}
	if caps.Scheduler != nil {
		if err := check("scheduler", caps.Scheduler.IP); err != nil {
			return err
		}
	}
	if caps.WebInterface != nil {
		if err := check("web_interface", caps.WebInterface.IP); err != nil {
			return err
		}
	}
	if caps.TimescaleDb != nil {
		if err := check("timescale_db", caps.TimescaleDb.IP); err != nil {
			return err
		}
	}
	if caps.Jupyter != nil {
		if err := check("jupyter", caps.Jupyter.IP); err != nil {
			return err
		}
	}

	return nil
}

// CheckProductionIPs ensures NO IP is 127.0.0.2
// -----------------------------------------------------------------------------

func CheckProductionIPs(config *core.Config) error {
	// Helper to check individual IP
	check := func(name, ip string) error {
		if ip == "127.0.0.2" {
			return fmt.Errorf("production integrity failure: %s IP cannot be 127.0.0.2 (Test IP detected)", name)
		}
		return nil
	}

	caps := config.Capabilities
	if caps.LogServer != nil {
		if err := check("log_server", caps.LogServer.IP); err != nil {
			return err
		}
	}
	if caps.ConfigServer != nil {
		if err := check("config_server", caps.ConfigServer.IP); err != nil {
			return err
		}
	}
	if caps.NotifServer != nil {
		if err := check("notif_server", caps.NotifServer.IP); err != nil {
			return err
		}
	}
	if caps.TeleRemote != nil {
		if err := check("tele_remote", caps.TeleRemote.IP); err != nil {
			return err
		}
	}
	if caps.Scheduler != nil {
		if err := check("scheduler", caps.Scheduler.IP); err != nil {
			return err
		}
	}
	if caps.WebInterface != nil {
		if err := check("web_interface", caps.WebInterface.IP); err != nil {
			return err
		}
	}
	if caps.TimescaleDb != nil {
		if err := check("timescale_db", caps.TimescaleDb.IP); err != nil {
			return err
		}
	}
	if caps.Jupyter != nil {
		if err := check("jupyter", caps.Jupyter.IP); err != nil {
			return err
		}
	}

	return nil
}
