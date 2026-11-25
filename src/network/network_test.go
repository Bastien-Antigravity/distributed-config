package network

// =============================================================================
// ESSENTIAL PROCESS:
// Unit test suite verifying ConfigProtoHandler message parsing, atomic LiveConfig
// store updates, and callback execution upon incoming network envelopes.
//
// DATA FLOW:
// 1. Input: Synthesized protobuf ConfigMessage payloads.
// 2. Logic: Ingests packets through handler.HandleData and asserts state changes.
// 3. Output: Test assertion results verifying callback execution and state updates.
//
// KEY PARAMETERS:
// - t: Standard testing harness handle.
// =============================================================================

import (
	"encoding/json"
	"errors"
	"math"
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"

	"google.golang.org/protobuf/proto"
)

// -----------------------------------------------------------------------------

func TestNetworkProtoHandler(t *testing.T) {
	config := &core.Config{}
	emptyMap := make(map[string]map[string]string)
	config.LiveConfig.Store(&emptyMap)

	handler := NewConfigHandler("TestHandler", config)

	t.Run("TestIncomingLiveConfigUpdate", func(t *testing.T) {
		// Mock callback trigger
		callbackTriggered := false
		handler.SetOnLiveConfUpdate(func(updates map[string]map[string]string) {
			callbackTriggered = true
			if updates["SECTION1"]["KEY1"] != "VAL1" {
				t.Errorf("Expected KEY1=VAL1, got %v", updates["SECTION1"]["KEY1"])
			}
		})

		payloadMap := map[string]map[string]string{
			"SECTION1": {
				"KEY1": "VAL1",
				"KEY2": "VAL2",
			},
		}
		dataPayload, _ := json.Marshal(payloadMap)

		// Create a mock PropagateLiveConfig message
		msg := &pb.ConfigMsg{
			Command: pb.ConfigMsg_BROADCAST_SYNC,
			Payload: dataPayload,
		}

		data, err := proto.Marshal(msg)
		if err != nil {
			t.Fatal(err)
		}

		// Handle Incoming
		if err := handler.HandleIncoming(data); err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}

		// Assertions
		if !callbackTriggered {
			t.Error("Expected callback to be triggered upon live config propagation")
		}

		if config.Get("SECTION1", "KEY1") != "VAL1" {
			t.Errorf("Expected LiveConfig to be updated, got %v", config.Get("SECTION1", "KEY1"))
		}
	})

	t.Run("TestIncomingRegistryUpdate", func(t *testing.T) {
		registryTriggered := false
		var capturedRegistry map[string][]string
		handler.SetOnRegistryUpdate(func(registry map[string][]string) {
			registryTriggered = true
			capturedRegistry = registry
		})

		payloadMap := map[string][]string{
			"active_services": {"service-a", "service-b"},
		}
		dataPayload, _ := json.Marshal(payloadMap)

		msg := &pb.ConfigMsg{
			Command: pb.ConfigMsg_BROADCAST_REGISTRY,
			Payload: dataPayload,
		}

		data, _ := proto.Marshal(msg)

		if err := handler.HandleIncoming(data); err != nil {
			t.Errorf("Expected success, got error: %v", err)
		}

		if !registryTriggered {
			t.Error("Expected registry callback to be triggered")
		}

		if len(capturedRegistry["active_services"]) != 2 {
			t.Errorf("Expected 2 services, got %d", len(capturedRegistry["active_services"]))
		}
	})

	t.Run("TestOutgoingRequests", func(t *testing.T) {
		data, err := handler.HandleOutgoing(pb.ConfigMsg_GET_SYNC, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(data) == 0 {
			t.Fatal("Expected non-empty wire payload for GET_SYNC so safe-socket does not treat frame as heartbeat")
		}

		msg := &pb.ConfigMsg{}
		if err := proto.Unmarshal(data, msg); err != nil {
			t.Fatal(err)
		}

		if msg.Command != pb.ConfigMsg_GET_SYNC {
			t.Errorf("Expected ReqClient get_live_config, got %v", msg.Command)
		}
	})

	t.Run("TestOutgoingUpdates", func(t *testing.T) {
		// Populate some live config to send
		config.Set(map[string]map[string]string{
			"OUTGOING": {"STATUS": "OK"},
		})

		data, err := handler.HandleOutgoing(pb.ConfigMsg_PUT_SYNC, nil) // passing nil defaults to LiveConfig
		if err != nil {
			t.Fatal(err)
		}

		msg := &pb.ConfigMsg{}
		if err := proto.Unmarshal(data, msg); err != nil {
			t.Fatal(err)
		}

		if msg.Command != pb.ConfigMsg_PUT_SYNC {
			t.Errorf("Expected ReqClient put_sync, got %v", msg.Command)
		}

		var decoded map[string]map[string]string
		if err := json.Unmarshal(msg.Payload, &decoded); err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		if decoded["OUTGOING"]["STATUS"] != "OK" {
			t.Errorf("Expected update payload to contain STATUS=OK")
		}
	})

	t.Run("TestGapDetection", func(t *testing.T) {
		gapHandler := NewConfigHandler("GapHandler", config)
		gapDetected := false
		var gotExpected, gotReceived uint64
		gapHandler.SetOnGapDetected(func(expected, received uint64) {
			gapDetected = true
			gotExpected = expected
			gotReceived = received
		})

		// Send version 1
		p1, _ := json.Marshal(map[string]map[string]string{"S": {"K": "1"}})
		m1, _ := proto.Marshal(&pb.ConfigMsg{Command: pb.ConfigMsg_BROADCAST_SYNC, Payload: p1, Version: 1})
		_ = gapHandler.HandleIncoming(m1)
		if gapHandler.GetLastSeenVersion() != 1 {
			t.Errorf("Expected version 1, got %d", gapHandler.GetLastSeenVersion())
		}

		// Send version 3 (gap! expected 2, got 3)
		p3, _ := json.Marshal(map[string]map[string]string{"S": {"K": "3"}})
		m3, _ := proto.Marshal(&pb.ConfigMsg{Command: pb.ConfigMsg_BROADCAST_SYNC, Payload: p3, Version: 3})
		_ = gapHandler.HandleIncoming(m3)

		if !gapDetected {
			t.Error("Expected gap detection callback to fire")
		}
		if gotExpected != 2 || gotReceived != 3 {
			t.Errorf("Expected gap (2, 3), got (%d, %d)", gotExpected, gotReceived)
		}
	})

	t.Run("TestOCC_OutgoingWithVersion", func(t *testing.T) {
		h := NewConfigHandler("OCCHandler", config)
		data, err := h.HandleOutgoingWithVersion(pb.ConfigMsg_PUT_SYNC, map[string]string{"K": "V"}, 42)
		if err != nil {
			t.Fatalf("HandleOutgoingWithVersion failed: %v", err)
		}
		var msg pb.ConfigMsg
		if err := proto.Unmarshal(data, &msg); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if msg.Version != 42 {
			t.Errorf("Expected msg.Version=42, got %d", msg.Version)
		}
	})

	t.Run("TestOCC_IncomingAckUpdatesVersion", func(t *testing.T) {
		h := NewConfigHandler("OCCHandler", config)
		var resultVer uint64
		var resultErr error
		h.SetOnUpdateResult(func(ver uint64, err error) {
			resultVer = ver
			resultErr = err
		})

		ackMsg, _ := proto.Marshal(&pb.ConfigMsg{
			Command: pb.ConfigMsg_ACK,
			Version: 88,
		})
		if err := h.HandleIncoming(ackMsg); err != nil {
			t.Fatalf("HandleIncoming failed: %v", err)
		}

		if h.GetLastSeenVersion() != 88 {
			t.Errorf("Expected lastSeenVersion=88, got %d", h.GetLastSeenVersion())
		}
		if resultVer != 88 || resultErr != nil {
			t.Errorf("Expected callback (88, nil), got (%d, %v)", resultVer, resultErr)
		}
	})

	t.Run("TestOCC_IncomingErrorMapsToErrVersionConflict", func(t *testing.T) {
		h := NewConfigHandler("OCCHandler", config)
		var resultVer uint64
		var resultErr error
		h.SetOnUpdateResult(func(ver uint64, err error) {
			resultVer = ver
			resultErr = err
		})

		errMsg, _ := proto.Marshal(&pb.ConfigMsg{
			Command: pb.ConfigMsg_ERROR,
			Payload: []byte("version mismatch conflict"),
			Version: 99,
		})
		err := h.HandleIncoming(errMsg)
		if err == nil {
			t.Fatal("Expected error, got nil")
		}
		if !errors.Is(err, ErrVersionConflict) {
			t.Errorf("Expected ErrVersionConflict, got %v", err)
		}
		if resultVer != 99 {
			t.Errorf("Expected callback ver=99, got %d", resultVer)
		}
		if h.GetLastSeenVersion() != 99 {
			t.Errorf("Expected lastSeenVersion=99, got %d", h.GetLastSeenVersion())
		}
		if !errors.Is(resultErr, ErrVersionConflict) {
			t.Errorf("Expected callback error ErrVersionConflict, got %v", resultErr)
		}
	})

	t.Run("TestDeletedSentinel_PreviewSet", func(t *testing.T) {
		cfg := &core.Config{}
		initial := map[string]map[string]string{
			"SEC": {
				"K1": "V1",
				"K2": "V2",
			},
		}
		cfg.LiveConfig.Store(&initial)

		// Delete K1 with sentinel
		res := cfg.PreviewSet(map[string]map[string]string{
			"SEC": {
				"K1": "__DELETED__",
			},
		})
		if (*res)["SEC"]["K1"] != "" {
			t.Errorf("Expected K1 to be deleted, got %v", (*res)["SEC"]["K1"])
		}
		if (*res)["SEC"]["K2"] != "V2" {
			t.Errorf("Expected K2 to survive, got %v", (*res)["SEC"]["K2"])
		}

		// Delete K2 with sentinel -> entire section SEC becomes empty and is removed
		res2 := cfg.PreviewSet(map[string]map[string]string{
			"SEC": {
				"K1": "__DELETED__",
				"K2": "__DELETED__",
			},
		})
		if _, ok := (*res2)["SEC"]; ok {
			t.Error("Expected empty section SEC to be purged")
		}
	})

	t.Run("TestConfigDelete", func(t *testing.T) {
		cfg := &core.Config{}
		initial := map[string]map[string]string{
			"SEC_A": {"K1": "V1", "K2": "V2"},
			"SEC_B": {"KB": "VB"},
		}
		cfg.LiveConfig.Store(&initial)

		// Delete specific key
		cfg.Delete(map[string][]string{"SEC_A": {"K1"}})
		live := *cfg.LiveConfig.Load()
		if _, ok := live["SEC_A"]["K1"]; ok {
			t.Error("Expected K1 in SEC_A to be deleted")
		}
		if live["SEC_A"]["K2"] != "V2" {
			t.Errorf("Expected K2 to be preserved, got %v", live["SEC_A"]["K2"])
		}

		// Delete entire section SEC_B by passing empty keys slice
		cfg.Delete(map[string][]string{"SEC_B": {}})
		live = *cfg.LiveConfig.Load()
		if _, ok := live["SEC_B"]; ok {
			t.Error("Expected SEC_B to be completely deleted")
		}
	})

	t.Run("TestStaleBroadcastSyncRejected", func(t *testing.T) {
		cfg := &core.Config{}
		initial := make(map[string]map[string]string)
		cfg.LiveConfig.Store(&initial)
		h := NewConfigHandler("StaleBroadcastHandler", cfg)

		// Apply version 5
		p5, _ := json.Marshal(map[string]map[string]string{"SEC": {"KEY": "V5"}})
		m5, _ := proto.Marshal(&pb.ConfigMsg{
			Command: pb.ConfigMsg_BROADCAST_SYNC,
			Payload: p5,
			Version: 5,
		})
		if err := h.HandleIncoming(m5); err != nil {
			t.Fatalf("HandleIncoming v5 failed: %v", err)
		}
		if h.GetLastSeenVersion() != 5 || cfg.Get("SEC", "KEY") != "V5" {
			t.Fatalf("Expected version 5 and KEY=V5, got ver=%d val=%q", h.GetLastSeenVersion(), cfg.Get("SEC", "KEY"))
		}

		// Send stale version 4 and duplicate version 5 -> both must be ignored
		for _, staleVer := range []int32{4, 5} {
			pStale, _ := json.Marshal(map[string]map[string]string{"SEC": {"KEY": "STALE"}})
			mStale, _ := proto.Marshal(&pb.ConfigMsg{
				Command: pb.ConfigMsg_BROADCAST_SYNC,
				Payload: pStale,
				Version: staleVer,
			})
			if err := h.HandleIncoming(mStale); err != nil {
				t.Fatalf("HandleIncoming stale v%d failed: %v", staleVer, err)
			}
			if h.GetLastSeenVersion() != 5 {
				t.Fatalf("Expected lastSeenVersion to remain 5 after v%d, got %d", staleVer, h.GetLastSeenVersion())
			}
			if cfg.Get("SEC", "KEY") != "V5" {
				t.Fatalf("Expected KEY to remain V5 after stale v%d, got %q", staleVer, cfg.Get("SEC", "KEY"))
			}
		}
	})

	t.Run("TestDelayedFullRefreshRejected", func(t *testing.T) {
		cfg := &core.Config{}
		initial := make(map[string]map[string]string)
		cfg.LiveConfig.Store(&initial)
		h := NewConfigHandler("StaleSnapshotHandler", cfg)

		// Apply delta broadcast at version 10
		p10, _ := json.Marshal(map[string]map[string]string{"SEC": {"KEY": "V10"}})
		m10, _ := proto.Marshal(&pb.ConfigMsg{
			Command: pb.ConfigMsg_BROADCAST_SYNC,
			Payload: p10,
			Version: 10,
		})
		if err := h.HandleIncoming(m10); err != nil {
			t.Fatalf("HandleIncoming v10 failed: %v", err)
		}

		// Delayed FULL_REFRESH at version 9 arrives -> must be rejected
		p9, _ := json.Marshal(map[string]map[string]string{"SEC": {"KEY": "V9_STALE"}})
		m9, _ := proto.Marshal(&pb.ConfigMsg{
			Command: pb.ConfigMsg_FULL_REFRESH,
			Payload: p9,
			Version: 9,
		})
		if err := h.HandleIncoming(m9); err != nil {
			t.Fatalf("HandleIncoming full refresh v9 failed: %v", err)
		}
		if h.GetLastSeenVersion() != 10 {
			t.Fatalf("Expected lastSeenVersion to stay 10, got %d", h.GetLastSeenVersion())
		}
		if cfg.Get("SEC", "KEY") != "V10" {
			t.Fatalf("Expected KEY to stay V10, got %q", cfg.Get("SEC", "KEY"))
		}

		// Newer FULL_REFRESH at version 10 or 11 must be accepted
		p11, _ := json.Marshal(map[string]map[string]string{"SEC": {"KEY": "V11_FRESH"}})
		m11, _ := proto.Marshal(&pb.ConfigMsg{
			Command: pb.ConfigMsg_FULL_REFRESH,
			Payload: p11,
			Version: 11,
		})
		if err := h.HandleIncoming(m11); err != nil {
			t.Fatalf("HandleIncoming full refresh v11 failed: %v", err)
		}
		if h.GetLastSeenVersion() != 11 || cfg.Get("SEC", "KEY") != "V11_FRESH" {
			t.Fatalf("Expected version 11 and V11_FRESH, got ver=%d val=%q", h.GetLastSeenVersion(), cfg.Get("SEC", "KEY"))
		}
		if cfg.GetVersion() != 11 {
			t.Fatalf("Expected parent Config.GetVersion() to stay synchronized at 11, got %d", cfg.GetVersion())
		}
	})

	t.Run("TestOutgoingVersionOverflowRejected", func(t *testing.T) {
		h := NewConfigHandler("OverflowHandler", config)
		_, err := h.HandleOutgoingWithVersion(pb.ConfigMsg_PUT_SYNC, map[string]map[string]string{"S": {"K": "V"}}, uint64(math.MaxInt32)+1)
		if err == nil {
			t.Fatal("Expected error when outgoing version exceeds protobuf int32 maximum")
		}
	})
}
