package core

// =============================================================================
// ESSENTIAL PROCESS:
// Unit and concurrency tests validating atomic pointer swapping (RCU), lock-free
// read consistency, and capability resolution in the core Config struct.
//
// DATA FLOW:
// 1. Input: Simulated configuration deltas and concurrent goroutine access.
// 2. Logic: Executes concurrent writes and reads against atomic.Pointer LiveConfig.
// 3. Output: Test assertion results verifying zero data races and read consistency.
//
// KEY PARAMETERS:
// - t: Standard testing harness context.
// =============================================================================

import (
	"sync"
	"testing"
)

// -----------------------------------------------------------------------------

func TestConfig_SetAndGet(t *testing.T) {
	cfg := &Config{}
	initialMap := make(map[string]map[string]string)
	cfg.LiveConfig.Store(&initialMap)

	t.Run("SingleUpdate", func(t *testing.T) {
		updates := map[string]map[string]string{
			"SECTION1": {"KEY1": "VAL1"},
		}
		cfg.Set(updates)

		val := cfg.Get("SECTION1", "KEY1")
		if val != "VAL1" {
			t.Errorf("Expected VAL1, got %s", val)
		}
	})

	t.Run("BulkUpdate_MultipleSections", func(t *testing.T) {
		updates := map[string]map[string]string{
			"SECTION1": {"KEY2": "VAL2"},
			"SECTION2": {"KEY3": "VAL3"},
		}
		cfg.Set(updates)

		if cfg.Get("SECTION1", "KEY1") != "VAL1" {
			t.Error("Previously set value SECTION1/KEY1 was lost")
		}
		if cfg.Get("SECTION1", "KEY2") != "VAL2" {
			t.Errorf("Expected VAL2, got %s", cfg.Get("SECTION1", "KEY2"))
		}
		if cfg.Get("SECTION2", "KEY3") != "VAL3" {
			t.Errorf("Expected VAL3, got %s", cfg.Get("SECTION2", "KEY3"))
		}
	})

	t.Run("NilUpdates_ShouldNotPanic", func(t *testing.T) {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("Set(nil) caused a panic: %v", r)
			}
		}()
		cfg.Set(nil)
	})
}

func TestConfig_ShareConfig(t *testing.T) {
	cfg := &Config{}
	initialMap := make(map[string]map[string]string)
	cfg.LiveConfig.Store(&initialMap)

	t.Run("ShareNestedMap", func(t *testing.T) {
		payload := map[string]map[string]string{
			"service_a": {"status": "up"},
		}
		err := cfg.ShareConfig(payload)
		if err != nil {
			t.Errorf("ShareConfig failed: %v", err)
		}
		if cfg.Get("service_a", "status") != "up" {
			t.Error("Nested map share failed")
		}
	})

	t.Run("ShareFlatMap_DefaultsToShared", func(t *testing.T) {
		payload := map[string]string{"ping": "pong"}
		err := cfg.ShareConfig(payload)
		if err != nil {
			t.Errorf("ShareConfig failed: %v", err)
		}
		if cfg.Get("shared", "ping") != "pong" {
			t.Error("Flat map share failed to use 'shared' section")
		}
	})

	t.Run("ShareInterfaceMap_Conversion", func(t *testing.T) {
		payload := map[string]interface{}{
			"version": 1.2,
			"active":  true,
		}
		err := cfg.ShareConfig(payload)
		if err != nil {
			t.Errorf("ShareConfig failed: %v", err)
		}
		if cfg.Get("shared", "version") != "1.2" {
			t.Errorf("Expected '1.2', got %s", cfg.Get("shared", "version"))
		}
		if cfg.Get("shared", "active") != "true" {
			t.Errorf("Expected 'true', got %s", cfg.Get("shared", "active"))
		}
	})
}

func TestConfig_ValidateMandatoryServices(t *testing.T) {
	t.Run("FailOnMissingServices", func(t *testing.T) {
		cfg := &Config{
			Capabilities: make(map[string]interface{}),
		}
		err := cfg.ValidateMandatoryServices()
		if err == nil {
			t.Error("Expected validation to fail on empty capabilities")
		}
	})

	t.Run("PassOnValidServices", func(t *testing.T) {
		cfg := &Config{
			Capabilities: map[string]interface{}{
				"log_server":    map[string]interface{}{"ip": "10.0.0.1", "port": "9000"},
				"config_server": map[string]interface{}{"ip": "10.0.0.1", "port": "3000"},
			},
		}
		err := cfg.ValidateMandatoryServices()
		if err != nil {
			t.Errorf("Expected validation to pass, got: %v", err)
		}
	})
}

func TestConfig_ShadowPort(t *testing.T) {
	cfg := &Config{
		Capabilities: map[string]interface{}{
			"my_service": map[string]interface{}{
				"ip":   "127.0.0.1",
				"port": "8000",
			},
			"grpc_fallback_service": map[string]interface{}{
				"ip":        "127.0.0.1",
				"port":      "8000",
				"grpc_port": "9998",
			},
			"explicit_service": map[string]interface{}{
				"ip":        "127.0.0.1",
				"port":      "8000",
				"grpc_ip":   "127.0.0.2",
				"grpc_port": "9999",
				"rest_port": "8080",
			},
			"explicit_rest_service": map[string]interface{}{
				"ip":        "127.0.0.1",
				"rest_ip":   "127.0.0.3",
				"rest_port": "8081",
			},
		},
	}
	initialMap := make(map[string]map[string]string)
	cfg.LiveConfig.Store(&initialMap)

	t.Run("MissingGRPCPortReturnsError", func(t *testing.T) {
		_, err := cfg.GetGRPCAddress("my_service")
		if err == nil {
			t.Fatalf("Expected error for missing grpc_port, got nil")
		}
	})

	t.Run("GRPCIPFallsBackToIP", func(t *testing.T) {
		addr, err := cfg.GetGRPCAddress("grpc_fallback_service")
		if err != nil {
			t.Fatalf("Expected success, got: %v", err)
		}
		if addr != "127.0.0.1:9998" {
			t.Errorf("Expected 127.0.0.1:9998, got %s", addr)
		}
	})

	t.Run("ExplicitGRPCIPAndPort", func(t *testing.T) {
		addr, err := cfg.GetGRPCAddress("explicit_service")
		if err != nil {
			t.Fatalf("Expected success, got: %v", err)
		}
		if addr != "127.0.0.2:9999" {
			t.Errorf("Expected 127.0.0.2:9999, got %s", addr)
		}
	})

	t.Run("RESTIPFallsBackToIP", func(t *testing.T) {
		addr, err := cfg.GetRESTAddress("explicit_service")
		if err != nil {
			t.Fatalf("Expected success, got: %v", err)
		}
		if addr != "127.0.0.1:8080" {
			t.Errorf("Expected 127.0.0.1:8080, got %s", addr)
		}
	})

	t.Run("ExplicitRESTIPAndPort", func(t *testing.T) {
		addr, err := cfg.GetRESTAddress("explicit_rest_service")
		if err != nil {
			t.Fatalf("Expected success, got: %v", err)
		}
		if addr != "127.0.0.3:8081" {
			t.Errorf("Expected 127.0.0.3:8081, got %s", addr)
		}
	})

	t.Run("MissingRESTPortReturnsError", func(t *testing.T) {
		_, err := cfg.GetRESTAddress("my_service")
		if err == nil {
			t.Fatalf("Expected error for missing rest_port, got nil")
		}
	})
}

func TestConfig_Concurrency(t *testing.T) {
	cfg := &Config{}
	initialMap := make(map[string]map[string]string)
	cfg.LiveConfig.Store(&initialMap)

	const workers = 10
	const iterations = 100
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				// Concurrent writes
				cfg.Set(map[string]map[string]string{
					"worker": {"last_id": string(rune(id))},
				})
				// Concurrent reads
				cfg.Get("worker", "last_id")
			}
		}(i)
	}

	wg.Wait()
}

// -----------------------------------------------------------------------------

func TestConfig_WildcardBindPreservedOverLiveConfig(t *testing.T) {
	cfg := &Config{
		Capabilities: map[string]interface{}{
			"notif_server": map[string]interface{}{
				"ip":        "0.0.0.0",
				"port":      "1026",
				"grpc_port": "1027",
				"rest_port": "1029",
			},
			"ipv6_service": map[string]interface{}{
				"ip":   "::",
				"port": "7000",
			},
			"peer_service": map[string]interface{}{
				"ip":        "127.0.0.1",
				"port":      "8000",
				"rest_port": "8080",
			},
		},
	}
	initialMap := make(map[string]map[string]string)
	cfg.LiveConfig.Store(&initialMap)

	cfg.Set(map[string]map[string]string{
		"notif_server": {
			"ip":      "notif-server",
			"rest_ip": "notif-server",
			"port":    "1026",
		},
		"ipv6_service": {
			"ip": "ipv6-dns",
		},
		"peer_service": {
			"ip": "peer-service",
		},
	})

	if addr, err := cfg.GetAddress("notif_server"); err != nil || addr != "0.0.0.0:1026" {
		t.Fatalf("Expected GetAddress to preserve wildcard 0.0.0.0:1026, got %q (err=%v)", addr, err)
	}
	if addr, err := cfg.GetGRPCAddress("notif_server"); err != nil || addr != "0.0.0.0:1027" {
		t.Fatalf("Expected GetGRPCAddress to preserve wildcard 0.0.0.0:1027, got %q (err=%v)", addr, err)
	}
	if addr, err := cfg.GetRESTAddress("notif_server"); err != nil || addr != "0.0.0.0:1029" {
		t.Fatalf("Expected GetRESTAddress to preserve wildcard 0.0.0.0:1029, got %q (err=%v)", addr, err)
	}
	if addr, err := cfg.GetAddress("ipv6_service"); err != nil || addr != ":::7000" {
		t.Fatalf("Expected GetAddress to preserve wildcard :::7000, got %q (err=%v)", addr, err)
	}
	if addr, err := cfg.GetAddress("peer_service"); err != nil || addr != "peer-service:8000" {
		t.Fatalf("Expected non-wildcard LiveConfig override to return peer-service:8000, got %q (err=%v)", addr, err)
	}
	if addr, err := cfg.GetRESTAddress("peer_service"); err != nil || addr != "peer-service:8080" {
		t.Fatalf("Expected non-wildcard REST LiveConfig override to return peer-service:8080, got %q (err=%v)", addr, err)
	}

	// Peer capability lookup via GetCapability must still reflect LiveConfig DNS override
	var notifCap struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
	}
	if err := cfg.GetCapability("notif_server", &notifCap); err != nil {
		t.Fatalf("GetCapability failed: %v", err)
	}
	if notifCap.IP != "notif-server" {
		t.Fatalf("Expected GetCapability to preserve LiveConfig peer IP 'notif-server', got %q", notifCap.IP)
	}
}

// -----------------------------------------------------------------------------

func TestConfig_ConcurrentWritersNoLostUpdates(t *testing.T) {
	cfg := &Config{}
	initialMap := make(map[string]map[string]string)
	cfg.LiveConfig.Store(&initialMap)

	const workers = 32
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			key := "key_" + string(rune('A'+id))
			cfg.Set(map[string]map[string]string{
				"concurrent_sec": {key: "present"},
			})
		}(i)
	}
	wg.Wait()

	for i := 0; i < workers; i++ {
		key := "key_" + string(rune('A'+i))
		if got := cfg.Get("concurrent_sec", key); got != "present" {
			t.Errorf("Lost update for %s: expected 'present', got %q", key, got)
		}
	}

	// Concurrent Delete of half the keys while setting new keys
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			key := "key_" + string(rune('A'+id))
			if id%2 == 0 {
				cfg.Delete(map[string][]string{"concurrent_sec": {key}})
			} else {
				cfg.Set(map[string]map[string]string{
					"concurrent_sec": {key: "updated"},
				})
			}
		}(i)
	}
	wg.Wait()

	for i := 0; i < workers; i++ {
		key := "key_" + string(rune('A'+i))
		got := cfg.Get("concurrent_sec", key)
		if i%2 == 0 && got != "" {
			t.Errorf("Expected key %s to be deleted, got %q", key, got)
		}
		if i%2 == 1 && got != "updated" {
			t.Errorf("Expected key %s to be 'updated', got %q", key, got)
		}
	}
}

// -----------------------------------------------------------------------------

func TestConfig_DynamicCapabilityGetAddress(t *testing.T) {
	cfg := &Config{
		Capabilities: map[string]interface{}{
			"static_service": map[string]interface{}{
				"ip":   "127.0.0.1",
				"port": "8000",
			},
		},
	}
	initialMap := make(map[string]map[string]string)
	cfg.LiveConfig.Store(&initialMap)

	// Unknown capability not in static or live must still fail
	if _, err := cfg.GetAddress("dynamic_service"); err == nil {
		t.Fatal("Expected error for unknown capability before live update")
	}

	// Add capability dynamically via LiveConfig section
	cfg.Set(map[string]map[string]string{
		"dynamic_service": {
			"ip":        "10.10.0.5",
			"port":      "9090",
			"grpc_port": "9091",
			"rest_port": "9092",
		},
		"capabilities": {
			"prefixed_service.ip":   "10.10.0.6",
			"prefixed_service.port": "7070",
		},
	})

	addr, err := cfg.GetAddress("dynamic_service")
	if err != nil || addr != "10.10.0.5:9090" {
		t.Fatalf("Expected dynamic_service address 10.10.0.5:9090, got %q (err=%v)", addr, err)
	}
	grpcAddr, err := cfg.GetGRPCAddress("dynamic_service")
	if err != nil || grpcAddr != "10.10.0.5:9091" {
		t.Fatalf("Expected dynamic_service gRPC address 10.10.0.5:9091, got %q (err=%v)", grpcAddr, err)
	}
	restAddr, err := cfg.GetRESTAddress("dynamic_service")
	if err != nil || restAddr != "10.10.0.5:9092" {
		t.Fatalf("Expected dynamic_service REST address 10.10.0.5:9092, got %q (err=%v)", restAddr, err)
	}
	prefAddr, err := cfg.GetAddress("prefixed_service")
	if err != nil || prefAddr != "10.10.0.6:7070" {
		t.Fatalf("Expected prefixed_service address 10.10.0.6:7070, got %q (err=%v)", prefAddr, err)
	}
}

// -----------------------------------------------------------------------------

func TestConfig_ResolveCapabilityAndVersion(t *testing.T) {
	cfg := NewDefaultConfig()
	cfg.Capabilities["db"] = map[string]interface{}{
		"ip":        "10.0.0.1",
		"port":      5432,
		"rest_port": 8080,
		"pool":      "10",
	}

	cfg.Set(map[string]map[string]string{
		"db": {
			"ip":   "10.0.0.2",
			"pool": "20",
		},
		"capabilities": {
			"db.pool": "30",
		},
	})

	resolved, err := cfg.ResolveCapability("db")
	if err != nil {
		t.Fatalf("ResolveCapability failed: %v", err)
	}
	if resolved["ip"] != "10.0.0.2" {
		t.Errorf("Expected direct LiveConfig ip 10.0.0.2, got %v", resolved["ip"])
	}
	if resolved["pool"] != "30" {
		t.Errorf("Expected dotted LiveConfig pool 30 to take precedence, got %v", resolved["pool"])
	}
	if resolved["port"] != "5432" || resolved["rest_port"] != "8080" {
		t.Errorf("Expected port/rest_port normalized to strings, got port=%T(%v) rest_port=%T(%v)",
			resolved["port"], resolved["port"], resolved["rest_port"], resolved["rest_port"])
	}

	// Mutating returned map must not affect subsequent calls
	resolved["pool"] = "mutated"
	resolved2, err := cfg.ResolveCapability("db")
	if err != nil {
		t.Fatalf("Second ResolveCapability failed: %v", err)
	}
	if resolved2["pool"] != "30" {
		t.Errorf("Expected detached map, got %v", resolved2["pool"])
	}

	if _, err := cfg.ResolveCapability("nonexistent"); err == nil {
		t.Fatal("Expected error for missing capability")
	}

	if cfg.GetVersion() != 0 {
		t.Errorf("Expected initial version 0, got %d", cfg.GetVersion())
	}
	cfg.SetVersion(42)
	if cfg.GetVersion() != 42 {
		t.Errorf("Expected version 42, got %d", cfg.GetVersion())
	}
}
