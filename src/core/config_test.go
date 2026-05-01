package core

import (
	"sync"
	"testing"
)

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
