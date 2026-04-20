package network

import (
	"encoding/json"
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"
	"google.golang.org/protobuf/proto"
)

func TestNetworkProtoHandler(t *testing.T) {
	config := &core.Config{
		LiveConfig: make(map[string]map[string]string),
	}
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

		if config.LiveConfig["SECTION1"]["KEY1"] != "VAL1" {
			t.Errorf("Expected LiveConfig to be updated, got %v", config.LiveConfig["SECTION1"])
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
		config.LiveConfig["OUTGOING"] = map[string]string{"STATUS": "OK"}

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
		json.Unmarshal(msg.Payload, &decoded)

		if decoded["OUTGOING"]["STATUS"] != "OK" {
			t.Errorf("Expected update payload to contain STATUS=OK")
		}
	})
}
