package strategies

// =============================================================================
// ESSENTIAL PROCESS:
// Strategy integration test suite verifying standalone fallback generation,
// cloud read-only vs writeback rules, and test environment IP enforcement.
//
// DATA FLOW:
// 1. Input: Temporary directory sandboxes, simulated configs, and mock network clients.
// 2. Logic: Invokes Strategy.Load and validates resulting config mutations.
// 3. Output: Test assertion results verifying strategy lifecycle correctness.
//
// KEY PARAMETERS:
// - t: Standard testing harness context.
// =============================================================================

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"testing"
	"time"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"

	safesocket "github.com/Bastien-Antigravity/safe-socket"

	"google.golang.org/protobuf/proto"
)

// -----------------------------------------------------------------------------

func TestStandaloneStrategy(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()

	t.Run("TestStandalone_FailsFastWhenFileMissing", func(t *testing.T) {
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldCwd) }()

		cfg := &core.Config{Logger: &strategyTestLogger{t: t}}
		strategy := &StandaloneStrategy{}

		err := strategy.Load(cfg)
		if err == nil {
			t.Errorf("Expected StandaloneStrategy Load to fail fast when standalone.yaml is missing")
		}
	})

	t.Run("TestStandalone_LoadsCanonicalStandaloneYAML", func(t *testing.T) {
		_ = os.Chdir(oldCwd)

		cfg := &core.Config{Logger: &strategyTestLogger{t: t}}
		strategy := &StandaloneStrategy{}

		err := strategy.Load(cfg)
		if err != nil {
			t.Errorf("Expected StandaloneStrategy Load to succeed from canonical standalone.yaml, got: %v", err)
		}

		if cfg.Common.Name != "common" {
			t.Errorf("Expected name 'common' from standalone.yaml, got: %v", cfg.Common.Name)
		}

		if err := cfg.ValidateMandatoryServices(); err != nil {
			t.Errorf("Expected canonical Standalone configuration to pass mandatory service validation, got: %v", err)
		}
	})
}

func TestCloudStrategy_Production(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(oldCwd) }()

	t.Run("TestProduction_FailsOnMissingEnvData", func(t *testing.T) {
		// In Production, missing files are ignored (returns nil).
		// However! Since no file provides config_server and environment
		// variables are not set, it will fail the ValidateMandatoryServices check at the end.
		cfg := &core.Config{Logger: &strategyTestLogger{t: t}}
		strategy := &CloudStrategy{Profile: "production", ReadOnly: false}

		err := strategy.Load(cfg)
		if err == nil {
			t.Errorf("Expected CloudStrategy (production) to fail-fast due to missing mandatory configuration (Log/Config Server)")
		}

		if cfg.Capabilities != nil && cfg.Capabilities["log_server"] != nil {
			t.Errorf("Should not have generated any capabilities for production missing file.")
		}
	})

	t.Run("TestProduction_IPSanityFailure", func(t *testing.T) {
		// Mock out a bad production file containing a test IP.
		_ = os.MkdirAll("config", 0755)
		badProdYaml := `
capabilities:
  log_server:
    ip: "127.0.0.2"
    port: "5000"
  config_server:
    ip: "192.168.1.100"
    port: "3306"
`
		_ = os.WriteFile("config/production.yaml", []byte(badProdYaml), 0644)

		cfg := &core.Config{Logger: &strategyTestLogger{t: t}}
		strategy := &CloudStrategy{Profile: "production", ReadOnly: false}

		err := strategy.Load(cfg)
		if err == nil {
			t.Errorf("Expected CloudStrategy (production) to fail due to CheckProductionIPs catching the 127.0.0.2 leak")
		}
	})
}

func TestCloudStrategy_Staging(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(oldCwd) }()

	t.Run("TestStaging_EnvironmentFirst_MissingDataFails", func(t *testing.T) {
		cfg := &core.Config{Logger: &strategyTestLogger{t: t}}
		strategy := &CloudStrategy{Profile: "staging", ReadOnly: true}

		err := strategy.Load(cfg)
		if err == nil {
			t.Errorf("Expected CloudStrategy (staging) to fail-fast due to missing mandatory infrastructure")
		}
	})
}

// -----------------------------------------------------------------------------

func TestTestStrategy(t *testing.T) {
	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()

	t.Run("TestTestStrategy_FailsFastWhenFileMissing", func(t *testing.T) {
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldCwd) }()

		cfg := &core.Config{Logger: &strategyTestLogger{t: t}}
		strategy := &TestStrategy{}

		err := strategy.Load(cfg)
		if err == nil {
			t.Errorf("Expected TestStrategy Load to fail fast when test.yaml is missing")
		}
	})

	t.Run("TestTestStrategy_SucceedsWithValidYAML", func(t *testing.T) {
		_ = os.Chdir(tempDir)
		defer func() { _ = os.Chdir(oldCwd) }()

		testYaml := `
common:
  name: "test-service"
capabilities:
  log_server:
    ip: "127.0.0.2"
    port: "5000"
  config_server:
    ip: "127.0.0.2"
    port: "3306"
`
		_ = os.WriteFile("test.yaml", []byte(testYaml), 0644)

		cfg := &core.Config{Logger: &strategyTestLogger{t: t}}
		strategy := &TestStrategy{}

		err := strategy.Load(cfg)
		if err != nil {
			t.Errorf("Expected TestStrategy Load to succeed with valid test config, got: %v", err)
		}
		if cfg.Common.Name != "test-service" {
			t.Errorf("Expected name 'test-service', got '%s'", cfg.Common.Name)
		}
		if cfg.Get("log_server", "ip") != "127.0.0.2" || cfg.Get("log_server", "port") != "5000" {
			t.Errorf("Expected LiveConfig to contain test.yaml capabilities, got ip=%q port=%q", cfg.Get("log_server", "ip"), cfg.Get("log_server", "port"))
		}
	})
}

// -----------------------------------------------------------------------------

func TestCloudStrategy_SyncsCapabilitiesToLiveConfigAndServer(t *testing.T) {
	// Find an ephemeral port on 127.0.0.1
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate ephemeral port: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	_ = ln.Close()
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	serverSock, err := safesocket.CreateWithConfig("tcp-hello", addr, safesocket.SocketConfig{
		Deadline: 5 * time.Second,
	}, "server", false)
	if err != nil {
		t.Fatalf("failed to create mock config-server socket: %v", err)
	}
	defer func() { _ = serverSock.Close() }()

	putSyncCh := make(chan map[string]map[string]string, 1)
	serverErrCh := make(chan error, 1)
	serverReady := make(chan struct{})
	go func() {
		if err := serverSock.Listen(); err != nil {
			serverErrCh <- fmt.Errorf("mock server Listen error: %w", err)
			return
		}
		close(serverReady)

		conn, err := serverSock.Accept()
		if err != nil {
			serverErrCh <- fmt.Errorf("mock server Accept error: %w", err)
			return
		}
		defer func() { _ = conn.Close() }()

		// Wait for the client's GET_SYNC frame before replying, proving end-to-end
		// that GET_SYNC produces a non-empty wire frame delivered by safe-socket.
		reqData, err := conn.ReadMessage()
		if err != nil {
			serverErrCh <- fmt.Errorf("mock server ReadMessage GET_SYNC error: %w", err)
			return
		}
		reqMsg := &pb.ConfigMsg{}
		if err := proto.Unmarshal(reqData, reqMsg); err != nil {
			serverErrCh <- fmt.Errorf("mock server Unmarshal GET_SYNC error: %w", err)
			return
		}
		if reqMsg.Command != pb.ConfigMsg_GET_SYNC {
			serverErrCh <- fmt.Errorf("expected client GET_SYNC command, got %v", reqMsg.Command)
			return
		}

		stale := map[string]map[string]string{
			"log_server": {"ip": "stale-ip", "port": "9999"},
			"shared":     {"keep": "yes"},
		}
		payload, _ := json.Marshal(stale)
		resp, _ := proto.Marshal(&pb.ConfigMsg{
			Command: pb.ConfigMsg_GET_SYNC,
			Payload: payload,
			Version: 1,
		})
		if _, err := conn.Write(resp); err != nil {
			serverErrCh <- fmt.Errorf("mock server Write GET_SYNC resp error: %w", err)
			return
		}

		for {
			data, err := conn.ReadMessage()
			if err != nil {
				return
			}
			msg := &pb.ConfigMsg{}
			if err := proto.Unmarshal(data, msg); err != nil {
				return
			}
			if msg.Command == pb.ConfigMsg_PUT_SYNC {
				var pushed map[string]map[string]string
				if err := json.Unmarshal(msg.Payload, &pushed); err == nil {
					putSyncCh <- pushed
				}
				return
			}
		}
	}()

	select {
	case <-serverReady:
	case err := <-serverErrCh:
		t.Fatalf("mock config-server error: %v", err)
	case <-time.After(2 * time.Second):
		t.Fatal("mock config-server failed to listen in time")
	}

	tempDir := t.TempDir()
	oldCwd, _ := os.Getwd()
	_ = os.Chdir(tempDir)
	defer func() { _ = os.Chdir(oldCwd) }()

	t.Setenv("DOCKER_ENV", "true")
	prodYaml := fmt.Sprintf(`
common:
  name: "cloud-test-service"
capabilities:
  config_server:
    ip: "127.0.0.1"
    port: "%d"
  log_server:
    ip: "127.0.0.2"
    port: 9020
    enabled: true
    empty_field: ""
    nested:
      ignored: "value"
`, port)
	_ = os.WriteFile("production.yaml", []byte(prodYaml), 0644)
	t.Setenv("CONFIG_PATH", "production.yaml")

	cfg := &core.Config{Logger: &strategyTestLogger{t: t}}
	strategy := &CloudStrategy{Profile: "production", ReadOnly: false}
	defer func() { _ = strategy.Close() }()

	if err := strategy.Load(cfg); err != nil {
		t.Fatalf("CloudStrategy.Load failed: %v", err)
	}

	if cfg.Get("log_server", "ip") != "127.0.0.2" {
		t.Fatalf("expected local file IP 127.0.0.2 to override stale server IP in LiveConfig, got %q", cfg.Get("log_server", "ip"))
	}
	if cfg.Get("log_server", "port") != "9020" {
		t.Fatalf("expected port 9020 in LiveConfig, got %q", cfg.Get("log_server", "port"))
	}
	if cfg.Get("log_server", "enabled") != "true" {
		t.Fatalf("expected bool scalar enabled=true in LiveConfig, got %q", cfg.Get("log_server", "enabled"))
	}
	if cfg.Get("log_server", "nested") != "" {
		t.Fatalf("expected nested map to be excluded from flat LiveConfig, got %q", cfg.Get("log_server", "nested"))
	}
	if cfg.Get("shared", "keep") != "yes" {
		t.Fatalf("expected server shared section to be preserved in LiveConfig, got %q", cfg.Get("shared", "keep"))
	}

	if err := strategy.Sync(cfg); err != nil {
		t.Fatalf("CloudStrategy.Sync failed: %v", err)
	}

	select {
	case pushed := <-putSyncCh:
		if pushed["log_server"]["ip"] != "127.0.0.2" || pushed["log_server"]["port"] != "9020" {
			t.Fatalf("expected PUT_SYNC to contain local log_server capabilities, got %+v", pushed["log_server"])
		}
		if pushed["config_server"]["ip"] != "127.0.0.1" {
			t.Fatalf("expected PUT_SYNC to contain config_server capabilities, got %+v", pushed["config_server"])
		}
	case err := <-serverErrCh:
		t.Fatalf("mock config-server error: %v", err)
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for PUT_SYNC from CloudStrategy.Sync")
	}
}

// -----------------------------------------------------------------------------

type strategyTestLogger struct {
	t *testing.T
}

func (l *strategyTestLogger) Debug(format string, args ...any)    { l.t.Logf("[DEBUG] "+format, args...) }
func (l *strategyTestLogger) Info(format string, args ...any)     { l.t.Logf("[INFO] "+format, args...) }
func (l *strategyTestLogger) Warning(format string, args ...any)  { l.t.Logf("[WARN] "+format, args...) }
func (l *strategyTestLogger) Error(format string, args ...any)    { l.t.Logf("[ERROR] "+format, args...) }
func (l *strategyTestLogger) Critical(format string, args ...any) { l.t.Logf("[CRIT] "+format, args...) }
func (l *strategyTestLogger) Logon(format string, args ...any)    {}
func (l *strategyTestLogger) Logout(format string, args ...any)   {}
func (l *strategyTestLogger) Trade(format string, args ...any)    {}
func (l *strategyTestLogger) Schedule(format string, args ...any) {}
func (l *strategyTestLogger) Report(format string, args ...any)   {}
func (l *strategyTestLogger) Stream(format string, args ...any)   {}
