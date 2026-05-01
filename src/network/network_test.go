package network

import (
	"encoding/json"
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"
	"google.golang.org/protobuf/proto"
)

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
}
