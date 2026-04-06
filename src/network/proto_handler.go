package network

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"

	"google.golang.org/protobuf/proto"
)

// -----------------------------------------------------------------------------

type ConfigProtoHandler struct {
	Name         string
	parentConfig *core.Config

	// Callbacks
	onMemConfUpdate   func(map[string]map[string]string)
	onRegistryUpdate  func(map[string][]string)
}

// -----------------------------------------------------------------------------

func NewConfigHandler(name string, config *core.Config) *ConfigProtoHandler {
	if name == "" {
		name = "ConfigProtoHandler"
	}
	return &ConfigProtoHandler{
		Name:         name,
		parentConfig: config,
	}
}

// Setters for callbacks
// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) SetOnMemConfUpdate(cb func(map[string]map[string]string)) {
	h.onMemConfUpdate = cb
}

func (h *ConfigProtoHandler) SetOnRegistryUpdate(cb func(map[string][]string)) {
	h.onRegistryUpdate = cb
}

// HandleOutgoing implements generic outgoing message creation
// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) HandleOutgoing(cmd pb.ConfigMsg_Cmd, payload interface{}) ([]byte, error) {
	var payloadBytes []byte
	var err error

	if payload != nil {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("json marshal error: %w", err)
		}
	} else if cmd == pb.ConfigMsg_PUT_SYNC && h.parentConfig.MemConfig != nil {
		// Default to sending current MemConfig
		payloadBytes, err = json.Marshal(h.parentConfig.MemConfig)
		if err != nil {
			return nil, err
		}
	}

	msg := &pb.ConfigMsg{
		Command: cmd,
		Payload: payloadBytes,
	}
	return proto.Marshal(msg)
}

// HandleIncoming processes incoming configurations and registries
// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) HandleIncoming(dataSer []byte) error {
	msg := &pb.ConfigMsg{}
	if err := proto.Unmarshal(dataSer, msg); err != nil {
		return fmt.Errorf("deserialization failed: %w", err)
	}

	switch msg.Command {
	case pb.ConfigMsg_BROADCAST_SYNC:
		var parsed map[string]map[string]string
		if err := json.Unmarshal(msg.Payload, &parsed); err != nil {
			return fmt.Errorf("failed to decode JSON payload: %w", err)
		}
		h.updateMemConfig(parsed)

	case pb.ConfigMsg_BROADCAST_REGISTRY:
		var parsed map[string][]string
		if err := json.Unmarshal(msg.Payload, &parsed); err != nil {
			return fmt.Errorf("failed to decode Registry JSON payload: %w", err)
		}
		if h.onRegistryUpdate != nil {
			h.onRegistryUpdate(parsed)
		}

	case pb.ConfigMsg_ACK:
		// No-op

	case pb.ConfigMsg_ERROR:
		return errors.New("server reported an error: " + string(msg.Payload))

	default:
		return fmt.Errorf("unknown server response command: %v", msg.Command)
	}
	return nil
}

// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) updateMemConfig(sections map[string]map[string]string) {
	if h.parentConfig.MemConfig == nil {
		h.parentConfig.MemConfig = make(map[string]map[string]string)
	}
	
	for sectKey, kv := range sections {
		h.parentConfig.MemConfig[sectKey] = kv
	}

	if h.onMemConfUpdate != nil {
		h.onMemConfUpdate(sections)
	}
}
