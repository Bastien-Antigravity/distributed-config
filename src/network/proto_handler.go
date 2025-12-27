package network

import (
	"errors"
	"fmt"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"

	"google.golang.org/protobuf/proto"
)

const COMMON_SECTION = "COMMON"

// -----------------------------------------------------------------------------

type ConfigProtoHandler struct {
	Name         string
	parentConfig *core.Config

	// Callbacks - aligned with legacy structure
	loggerRefreshLogLevel func(map[string][]string)
	notifRefreshSender    func(map[string]map[string]string) map[string][]string
	loggerLog             func(string, string)
}

// ConnectionHandler interface support removed

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

func (h *ConfigProtoHandler) SetLoggerCallBack(cb func(map[string][]string)) {
	h.loggerRefreshLogLevel = cb
}

// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) SetNotifCallBack(cb func(map[string]map[string]string) map[string][]string) {
	h.notifRefreshSender = cb
}

// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) SetLoggerLog(cb func(string, string)) {
	h.loggerLog = cb
}

// HandleOutgoing implements ConnectionHandler
// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) HandleOutgoing(cmd interface{}, payload interface{}) ([]byte, error) {
	clientCmd, ok := cmd.(pb.ConfigMsg_ConfigClientCmd)
	if !ok {
		return nil, fmt.Errorf("invalid command type: %T", cmd)
	}

	switch clientCmd {
	case pb.ConfigMsg_update_mem_config:
		sectionMap := map[string]*pb.KeysValues{}
		if h.parentConfig.MemConfig != nil {
			for sectKey, keyValMap := range h.parentConfig.MemConfig {
				if len(keyValMap) > 0 {
					cleanMap := make(map[string]string)
					for key, val := range keyValMap {
						if val != "" {
							cleanMap[key] = val
						}
					}
					if len(cleanMap) > 0 {
						sectionMap[sectKey] = &pb.KeysValues{KeyValue: cleanMap}
					}
				}
			}
		}
		msg := &pb.ConfigMsg{
			ReqClient:          pb.ConfigMsg_update_mem_config,
			SectionsKeysValues: sectionMap,
		}
		return proto.Marshal(msg)

	case pb.ConfigMsg_update_config_object:
		// Logic adapted from legacy:
		// for _, section := range ConfigProtoHandler.parentClassConfig.Parser.Sections() ...
		// core.Config doesn't expose Parser sections directly in the same way (it's pure data).
		// We'll skip for now or mapping would require more infrastructure in core.Config
		sectionMap := map[string]*pb.KeysValues{}
		msg := &pb.ConfigMsg{
			ReqClient:          pb.ConfigMsg_update_config_object,
			SectionsKeysValues: sectionMap,
		}
		return proto.Marshal(msg)

	case pb.ConfigMsg_get_mem_config:
		msg := &pb.ConfigMsg{ReqClient: pb.ConfigMsg_get_mem_config}
		return proto.Marshal(msg)

	case pb.ConfigMsg_get_config_object:
		msg := &pb.ConfigMsg{ReqClient: pb.ConfigMsg_get_config_object}
		return proto.Marshal(msg)

	case pb.ConfigMsg_add_config_listener:
		msg := &pb.ConfigMsg{ReqClient: pb.ConfigMsg_add_config_listener}
		return proto.Marshal(msg)

	case pb.ConfigMsg_dump_mem_config:
		msg := &pb.ConfigMsg{ReqClient: pb.ConfigMsg_dump_mem_config}
		return proto.Marshal(msg)

	case pb.ConfigMsg_get_notif_loglevel:
		msg := &pb.ConfigMsg{ReqClient: pb.ConfigMsg_get_notif_loglevel}
		return proto.Marshal(msg)

	case pb.ConfigMsg_update_notif_loglevel:
		sectionMap := map[string]*pb.KeysValues{}
		// Skipping async callback logic for simplicity in port, consistent with previous attempt.
		msg := &pb.ConfigMsg{
			ReqClient:          pb.ConfigMsg_update_mem_config, // Legacy uses update_mem_config here?? Verified in legacy code.
			SectionsKeysValues: sectionMap,
		}
		return proto.Marshal(msg)

	default:
		return nil, fmt.Errorf("unknown client cmd: %v", clientCmd)
	}
}

// HandleIncoming implements ConnectionHandler
// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) HandleIncoming(dataSer []byte) error {
	msg := &pb.ConfigMsg{}
	if err := proto.Unmarshal(dataSer, msg); err != nil {
		return fmt.Errorf("deserialization failed: %w", err)
	}

	switch msg.RespServer {
	case pb.ConfigMsg_propagate_mem_config:
		h.updateMemConfig(msg.SectionsKeysValues)

	case pb.ConfigMsg_mem_config_update_done:
		// No-op

	case pb.ConfigMsg_propagate_config:
		// Logic adapted:
		// if sectKey == COMMON_SECTION { parent.UpdateSelf(...) }
		// parent.Parser.Section(...).MapTo(...)
		// Core Config is struct based.
		// We can support Common update manually.
		// for sectKey, keysValues := range msg.SectionsKeysValues {
		// 	if sectKey == COMMON_SECTION {
		// 		// We don't have UpdateSelf on core.Config (it was on Facade wrapper?)
		// 		// We'll update CommonConfig directly if possible or map it.
		// 		// h.mapCommon(...)
		// 	}
		// }

	case pb.ConfigMsg_propagate_notif_loglevel:
		// h.handleNotifLogLevel(msg.SectionsKeysValues)

	case pb.ConfigMsg_send_mem_config_init:
		h.updateMemConfig(msg.SectionsKeysValues)

	case pb.ConfigMsg_send_config_init:
		// Same as propagate_config

	case pb.ConfigMsg_send_notif_loglevel_init:
		// h.handleNotifLogLevel(msg.SectionsKeysValues)

	case pb.ConfigMsg_mem_config_update_failed:
		return errors.New("server reported mem_config_update_failed")

	case pb.ConfigMsg_config_update_failed:
		return errors.New("server reported config_update_failed")

	default:
		return fmt.Errorf("unknown server response: %s", msg.RespServer)
	}
	return nil
}

// -----------------------------------------------------------------------------

func (h *ConfigProtoHandler) updateMemConfig(sections map[string]*pb.KeysValues) {
	if h.parentConfig.MemConfig == nil {
		h.parentConfig.MemConfig = make(map[string]map[string]string)
	}
	for sectKey, kv := range sections {
		h.parentConfig.MemConfig[sectKey] = kv.GetKeyValue()
	}
}
