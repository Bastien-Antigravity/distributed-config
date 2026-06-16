package network

import (
	"encoding/json"
	"testing"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"
	"google.golang.org/protobuf/proto"
)

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
