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
	"math"
	"strings"
	"sync"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"

	"google.golang.org/protobuf/proto"
)

// -----------------------------------------------------------------------------

// ErrVersionConflict is returned when an OCC update or delete fails due to a version mismatch.
var ErrVersionConflict = errors.New("concurrent modification detected (version mismatch)")

// -----------------------------------------------------------------------------

type ConfigProtoHandler struct {
	Name         string
	parentConfig *core.Config

	// Callbacks
	onLiveConfUpdate func(map[string]map[string]string)
	onRegistryUpdate func(map[string][]string)
	onSyncReceived   func()
	onUpdateResult   func(newVersion uint64, err error)
	onGapDetected    func(expected, received uint64)
	lastSeenVersion  uint64
	lastAckedVersion uint64
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
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onLiveConfUpdate = cb
}

func (h *ConfigProtoHandler) SetOnRegistryUpdate(cb func(map[string][]string)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onRegistryUpdate = cb
}

func (h *ConfigProtoHandler) SetOnSyncReceived(cb func()) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onSyncReceived = cb
}

func (h *ConfigProtoHandler) SetOnGapDetected(cb func(expected, received uint64)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onGapDetected = cb
}

func (h *ConfigProtoHandler) SetOnUpdateResult(cb func(newVersion uint64, err error)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onUpdateResult = cb
}

func (h *ConfigProtoHandler) GetLastSeenVersion() uint64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastSeenVersion
}

// ResetSessionVersion clears session-local version tracking on TCP reconnect so a
// restarted config-server with a lower store revision can be resynchronized via GET_SYNC.
func (h *ConfigProtoHandler) ResetSessionVersion() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastSeenVersion = 0
	h.lastAckedVersion = 0
}

// HandleOutgoing implements generic outgoing message creation
// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) HandleOutgoing(cmd pb.ConfigMsg_Cmd, payload interface{}) ([]byte, error) {
	return h.HandleOutgoingWithVersion(cmd, payload, 0)
}

// HandleOutgoingWithVersion builds outgoing message with an explicit version for OCC
// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) HandleOutgoingWithVersion(cmd pb.ConfigMsg_Cmd, payload interface{}, version uint64) ([]byte, error) {
	if version > math.MaxInt32 {
		return nil, fmt.Errorf("version %d exceeds protobuf int32 maximum", version)
	}

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

	if cmd == pb.ConfigMsg_GET_SYNC && len(payloadBytes) == 0 {
		// Proto3 marshals GET_SYNC (enum 0) with empty fields to 0 bytes, which safe-socket
		// FramedTCPSocket treats as a heartbeat frame and skips.
		payloadBytes = []byte("{}")
	}

	msg := &pb.ConfigMsg{
		Command: cmd,
		Payload: payloadBytes,
		Version: int32(version),
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

		// Monotonic version verification & sequence gap detection
		if msg.Version > 0 {
			incomingVer := uint64(msg.Version)
			h.mu.Lock()
			if h.lastSeenVersion > 0 && incomingVer <= h.lastSeenVersion {
				if incomingVer == h.lastAckedVersion {
					// Own mutation ACK arrived first; allow the matching broadcast through once so onLiveConfUpdate fires once.
					h.lastAckedVersion = 0
				} else {
					h.mu.Unlock()
					return nil
				}
			}
			if h.lastSeenVersion > 0 && incomingVer > h.lastSeenVersion+1 {
				gapCb := h.onGapDetected
				expected := h.lastSeenVersion + 1
				h.lastSeenVersion = incomingVer
				h.lastAckedVersion = 0
				if h.parentConfig != nil {
					h.parentConfig.SetVersion(incomingVer)
				}
				h.mu.Unlock()
				if gapCb != nil {
					gapCb(expected, incomingVer)
				}
			} else {
				if incomingVer > h.lastSeenVersion {
					h.lastSeenVersion = incomingVer
					h.lastAckedVersion = 0
					if h.parentConfig != nil {
						h.parentConfig.SetVersion(incomingVer)
					}
				}
				h.mu.Unlock()
			}
		}

		// Use Set() for merging (Delta Update)
		h.parentConfig.Set(parsed)
		h.mu.RLock()
		liveCb := h.onLiveConfUpdate
		h.mu.RUnlock()
		if liveCb != nil {
			liveCb(parsed)
		}

	case pb.ConfigMsg_BROADCAST_REGISTRY:
		var parsed map[string][]string
		if err := json.Unmarshal(msg.Payload, &parsed); err != nil {
			return fmt.Errorf("failed to decode Registry JSON payload: %w", err)
		}
		h.mu.RLock()
		regCb := h.onRegistryUpdate
		h.mu.RUnlock()
		if regCb != nil {
			regCb(parsed)
		}

	case pb.ConfigMsg_ACK:
		if msg.Version > 0 {
			h.mu.Lock()
			if uint64(msg.Version) > h.lastSeenVersion {
				h.lastSeenVersion = uint64(msg.Version)
				h.lastAckedVersion = uint64(msg.Version)
				if h.parentConfig != nil {
					h.parentConfig.SetVersion(h.lastSeenVersion)
				}
			}
			h.mu.Unlock()
		}
		h.mu.RLock()
		updateCb := h.onUpdateResult
		h.mu.RUnlock()
		if updateCb != nil {
			updateCb(uint64(msg.Version), nil)
		}

	case pb.ConfigMsg_GET_SYNC, pb.ConfigMsg_FULL_REFRESH: // Added direct sync support
		var parsed map[string]map[string]string
		if err := json.Unmarshal(msg.Payload, &parsed); err != nil {
			return fmt.Errorf("failed to decode GET_SYNC/FULL_REFRESH JSON payload: %w", err)
		}
		if msg.Version > 0 {
			incomingVer := uint64(msg.Version)
			h.mu.Lock()
			if h.lastSeenVersion > 0 && incomingVer < h.lastSeenVersion {
				syncCb := h.onSyncReceived
				h.mu.Unlock()
				if syncCb != nil {
					syncCb()
				}
				return nil
			}
			h.lastSeenVersion = incomingVer
			h.lastAckedVersion = 0
			if h.parentConfig != nil {
				h.parentConfig.SetVersion(incomingVer)
			}
			h.mu.Unlock()
		}
		h.updateLiveConfig(parsed)
		h.mu.RLock()
		syncCb := h.onSyncReceived
		h.mu.RUnlock()
		if syncCb != nil {
			syncCb()
		}

	case pb.ConfigMsg_ERROR:
		msgStr := string(msg.Payload)
		var err error
		if strings.Contains(strings.ToLower(msgStr), "version mismatch") {
			err = fmt.Errorf("%w: %s", ErrVersionConflict, msgStr)
		} else {
			err = errors.New("server reported an error: " + msgStr)
		}
		if msg.Version > 0 {
			h.mu.Lock()
			if uint64(msg.Version) > h.lastSeenVersion {
				h.lastSeenVersion = uint64(msg.Version)
				if h.parentConfig != nil {
					h.parentConfig.SetVersion(h.lastSeenVersion)
				}
			}
			h.mu.Unlock()
		}
		h.mu.RLock()
		updateCb := h.onUpdateResult
		h.mu.RUnlock()
		if updateCb != nil {
			updateCb(uint64(msg.Version), err)
		}
		return err

	default:
		return fmt.Errorf("unknown server response command: %v", msg.Command)
	}
	return nil
}

// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) updateLiveConfig(sections map[string]map[string]string) {
	// Atomically swap the entire config map pointer serialized under parentConfig's writer lock
	h.parentConfig.ReplaceLive(sections)

	h.mu.RLock()
	liveCb := h.onLiveConfUpdate
	h.mu.RUnlock()
	if liveCb != nil {
		liveCb(sections)
	}
}
