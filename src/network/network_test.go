package network

import (
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"
	"google.golang.org/protobuf/proto"
)

func TestNetworkProtoHandler(t *testing.T) {
	config := &core.Config{
		MemConfig: make(map[string]map[string]string),
	}
	handler := NewConfigHandler("TestHandler", config)

	t.Run("TestIncomingMemConfigUpdate", func(t *testing.T) {
		// Mock callback trigger
		callbackTriggered := false
		handler.SetOnMemConfUpdate(func(updates map[string]map[string]string) {
			callbackTriggered = true
			if updates["SECTION1"]["KEY1"] != "VAL1" {
				t.Errorf("Expected KEY1=VAL1, got %v", updates["SECTION1"]["KEY1"])
			}
		})

		// Create a mock PropagateMemConfig message
		msg := &pb.ConfigMsg{
			RespServer: pb.ConfigMsg_propagate_mem_config,
			SectionsKeysValues: map[string]*pb.KeysValues{
				"SECTION1": {
					KeyValue: map[string]string{
						"KEY1": "VAL1",
						"KEY2": "VAL2",
					},
				},
			},
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
			t.Error("Expected callback to be triggered upon mem config propagation")
		}

		if config.MemConfig["SECTION1"]["KEY1"] != "VAL1" {
			t.Errorf("Expected MemConfig to be updated, got %v", config.MemConfig["SECTION1"])
		}
	})

	t.Run("TestOutgoingRequests", func(t *testing.T) {
		// Test get_mem_config request generation
		data, err := handler.HandleOutgoing(pb.ConfigMsg_get_mem_config, nil)
		if err != nil {
			t.Fatal(err)
		}

		msg := &pb.ConfigMsg{}
		if err := proto.Unmarshal(data, msg); err != nil {
			t.Fatal(err)
		}

		if msg.ReqClient != pb.ConfigMsg_get_mem_config {
			t.Errorf("Expected ReqClient get_mem_config, got %v", msg.ReqClient)
		}
	})

	t.Run("TestOutgoingUpdates", func(t *testing.T) {
		// Populate some mem config to send
		config.MemConfig["OUTGOING"] = map[string]string{"STATUS": "OK"}

		data, err := handler.HandleOutgoing(pb.ConfigMsg_update_mem_config, nil)
		if err != nil {
			t.Fatal(err)
		}

		msg := &pb.ConfigMsg{}
		if err := proto.Unmarshal(data, msg); err != nil {
			t.Fatal(err)
		}

		if msg.ReqClient != pb.ConfigMsg_update_mem_config {
			t.Errorf("Expected ReqClient update_mem_config, got %v", msg.ReqClient)
		}

		if msg.SectionsKeysValues["OUTGOING"].KeyValue["STATUS"] != "OK" {
			t.Errorf("Expected update payload to contain STATUS=OK, got %v", msg.SectionsKeysValues["OUTGOING"])
		}
	})
}
