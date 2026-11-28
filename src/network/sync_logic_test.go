package network

// =============================================================================
// ESSENTIAL PROCESS:
// Regression and sync logic test suite verifying that broadcast synchronization
// deep-merges delta updates into LiveConfig without dropping untouched sections.
//
// DATA FLOW:
// 1. Input: Multi-section initial LiveConfig state and simulated delta updates.
// 2. Logic: Feeds updates through ConfigProtoHandler and evaluates resulting atomic map.
// 3. Output: Test assertion results verifying non-destructive merge semantics.
//
// KEY PARAMETERS:
// - t: Testing harness handle.
// =============================================================================

import (
	"encoding/json"
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"
	"google.golang.org/protobuf/proto"
)

// -----------------------------------------------------------------------------

func TestMergeBugReproduction(t *testing.T) {
	config := &core.Config{}
	initialMap := map[string]map[string]string{
		"SECTION_A": {"KEY1": "VAL1"},
		"SECTION_B": {"KEY2": "VAL2"},
	}
	config.LiveConfig.Store(&initialMap)

	handler := NewConfigHandler("TestHandler", config)

	t.Run("BROADCAST_SYNC_ShouldMergeNotReplace", func(t *testing.T) {
		deltaMap := map[string]map[string]string{
			"SECTION_A": {"KEY1": "UPDATED"},
		}
		dataPayload, _ := json.Marshal(deltaMap)

		msg := &pb.ConfigMsg{
			Command: pb.ConfigMsg_BROADCAST_SYNC,
			Payload: dataPayload,
		}

		data, _ := proto.Marshal(msg)

		if err := handler.HandleIncoming(data); err != nil {
			t.Fatalf("HandleIncoming failed: %v", err)
		}

		// Check if SECTION_A.KEY1 is updated
		if config.Get("SECTION_A", "KEY1") != "UPDATED" {
			t.Errorf("Expected UPDATED, got %v", config.Get("SECTION_A", "KEY1"))
		}

		// Check if SECTION_B.KEY2 is STILL THERE
		if config.Get("SECTION_B", "KEY2") != "VAL2" {
			t.Errorf("BUG: SECTION_B was lost! Expected VAL2, got %v", config.Get("SECTION_B", "KEY2"))
		}
	})

	t.Run("GET_SYNC_ShouldReplace", func(t *testing.T) {
		fullMap := map[string]map[string]string{
			"SECTION_C": {"KEY3": "VAL3"},
		}
		dataPayload, _ := json.Marshal(fullMap)

		msg := &pb.ConfigMsg{
			Command: pb.ConfigMsg_GET_SYNC,
			Payload: dataPayload,
		}

		data, _ := proto.Marshal(msg)

		if err := handler.HandleIncoming(data); err != nil {
			t.Fatalf("HandleIncoming failed: %v", err)
		}

		// SECTION_A and SECTION_B should be GONE
		if config.Get("SECTION_A", "KEY1") != "" {
			t.Error("Expected SECTION_A to be gone after GET_SYNC")
		}
		if config.Get("SECTION_C", "KEY3") != "VAL3" {
			t.Errorf("Expected VAL3, got %v", config.Get("SECTION_C", "KEY3"))
		}
	})

	t.Run("FULL_REFRESH_ShouldReplace", func(t *testing.T) {
		fullMap := map[string]map[string]string{
			"SECTION_D": {"KEY4": "VAL4"},
		}
		dataPayload, _ := json.Marshal(fullMap)

		msg := &pb.ConfigMsg{
			Command: pb.ConfigMsg_FULL_REFRESH,
			Payload: dataPayload,
		}

		data, _ := proto.Marshal(msg)

		if err := handler.HandleIncoming(data); err != nil {
			t.Fatalf("HandleIncoming failed: %v", err)
		}

		// SECTION_C should be GONE
		if config.Get("SECTION_C", "KEY3") != "" {
			t.Error("Expected SECTION_C to be gone after FULL_REFRESH")
		}
		if config.Get("SECTION_D", "KEY4") != "VAL4" {
			t.Errorf("Expected VAL4, got %v", config.Get("SECTION_D", "KEY4"))
		}
	})
}

// -----------------------------------------------------------------------------

func TestProtoHandler_ServerRestartResetAndAckDedup(t *testing.T) {
	cfg := &core.Config{}
	liveCount := 0
	handler := NewConfigHandler("RestartHandler", cfg)
	handler.SetOnLiveConfUpdate(func(_ map[string]map[string]string) {
		liveCount++
	})

	// 1. Client reaches version 15 from initial server session.
	snap15, _ := json.Marshal(map[string]map[string]string{"svc": {"port": "8080"}})
	raw15, _ := proto.Marshal(&pb.ConfigMsg{Command: pb.ConfigMsg_GET_SYNC, Version: 15, Payload: snap15})
	if err := handler.HandleIncoming(raw15); err != nil {
		t.Fatalf("HandleIncoming(v15) failed: %v", err)
	}
	if handler.GetLastSeenVersion() != 15 {
		t.Fatalf("Expected lastSeenVersion 15, got %d", handler.GetLastSeenVersion())
	}

	// 2. TCP reconnect calls ResetSessionVersion before GET_SYNC; server restart with lower version 2 is adopted.
	handler.ResetSessionVersion()
	snap2, _ := json.Marshal(map[string]map[string]string{"svc": {"port": "9090"}})
	raw2, _ := proto.Marshal(&pb.ConfigMsg{Command: pb.ConfigMsg_GET_SYNC, Version: 2, Payload: snap2})
	if err := handler.HandleIncoming(raw2); err != nil {
		t.Fatalf("HandleIncoming(v2) failed: %v", err)
	}
	if handler.GetLastSeenVersion() != 2 {
		t.Fatalf("Expected lastSeenVersion 2 after restart snapshot, got %d", handler.GetLastSeenVersion())
	}
	if cfg.Get("svc", "port") != "9090" {
		t.Fatalf("Expected port 9090 from restart snapshot, got %q", cfg.Get("svc", "port"))
	}

	// 3. Subsequent ACK at v3 followed by BROADCAST_SYNC at v3 fires onLiveConfUpdate once,
	//    and any duplicate BROADCAST_SYNC at v3 is ignored.
	beforeLive := liveCount
	ackRaw, _ := proto.Marshal(&pb.ConfigMsg{Command: pb.ConfigMsg_ACK, Version: 3})
	if err := handler.HandleIncoming(ackRaw); err != nil {
		t.Fatalf("HandleIncoming(ACK v3) failed: %v", err)
	}
	bcastPayload, _ := json.Marshal(map[string]map[string]string{"svc": {"port": "9091"}})
	bcastRaw, _ := proto.Marshal(&pb.ConfigMsg{Command: pb.ConfigMsg_BROADCAST_SYNC, Version: 3, Payload: bcastPayload})
	if err := handler.HandleIncoming(bcastRaw); err != nil {
		t.Fatalf("HandleIncoming(BROADCAST v3 #1) failed: %v", err)
	}
	if err := handler.HandleIncoming(bcastRaw); err != nil {
		t.Fatalf("HandleIncoming(BROADCAST v3 #2) failed: %v", err)
	}
	if liveCount != beforeLive+1 {
		t.Fatalf("Expected exactly one onLiveConfUpdate for ACK+duplicate BROADCAST, got %d -> %d", beforeLive, liveCount)
	}
}
