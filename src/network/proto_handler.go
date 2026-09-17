package network

// =============================================================================
// ESSENTIAL PROCESS:
// Protobuf protocol handler deserializing incoming wire messages from config-server,
// updating local atomic configuration snapshots, and triggering client callbacks.
//
// DATA FLOW:
// 1. Input: Binary serialized protobuf payloads (ConfigMessage envelope).
// 2. Logic: Demultiplexes payload types (InitialConfig, LiveUpdate, ServiceRegistryUpdate).
// 3. Output: Dispatches parsed configurations to parent Config and triggers registered callbacks.
//
// KEY PARAMETERS:
// - Name: Identifier of the handler instance.
// - onLiveConfUpdate: Callback triggered on dynamic config modifications.
// =============================================================================

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"

	"google.golang.org/protobuf/proto"
)

// -----------------------------------------------------------------------------

type ConfigProtoHandler struct {
	Name         string
	parentConfig *core.Config

	// Callbacks
	onLiveConfUpdate func(map[string]map[string]string)
	onRegistryUpdate func(map[string][]string)
	onSyncReceived   func()
	mu               sync.RWMutex
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

func (h *ConfigProtoHandler) SetOnLiveConfUpdate(cb func(map[string]map[string]string)) {
	h.onLiveConfUpdate = cb
}

func (h *ConfigProtoHandler) SetOnRegistryUpdate(cb func(map[string][]string)) {
	h.onRegistryUpdate = cb
}

func (h *ConfigProtoHandler) SetOnSyncReceived(cb func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onSyncReceived = cb
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
	} else if cmd == pb.ConfigMsg_PUT_SYNC && h.parentConfig.LiveConfig.Load() != nil {
		// Default to sending current LiveConfig snapshot
		payloadBytes, err = json.Marshal(h.parentConfig.LiveConfig.Load())
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
		// Use Set() for merging (Delta Update)
		h.parentConfig.Set(parsed)
		if h.onLiveConfUpdate != nil {
			h.onLiveConfUpdate(parsed)
		}

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

	case pb.ConfigMsg_GET_SYNC, pb.ConfigMsg_FULL_REFRESH: // Added direct sync support
		var parsed map[string]map[string]string
		if err := json.Unmarshal(msg.Payload, &parsed); err != nil {
			return fmt.Errorf("failed to decode GET_SYNC/FULL_REFRESH JSON payload: %w", err)
		}
		h.updateLiveConfig(parsed)
		h.mu.RLock()
		syncCb := h.onSyncReceived
		h.mu.RUnlock()
		if syncCb != nil {
			syncCb()
		}

	case pb.ConfigMsg_ERROR:
		return errors.New("server reported an error: " + string(msg.Payload))

	default:
		return fmt.Errorf("unknown server response command: %v", msg.Command)
	}
	return nil
}

// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) updateLiveConfig(sections map[string]map[string]string) {
	// Atomically swap the entire config map pointer (Full Update from Server)
	h.parentConfig.LiveConfig.Store(&sections)

	if h.onLiveConfUpdate != nil {
		h.onLiveConfUpdate(sections)
	}
}
