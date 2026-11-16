package network

// =============================================================================
// ESSENTIAL PROCESS:
// safe-socket network client managing TCP connection lifecycle, exponential
// backoff reconnection, and real-time bidirectional synchronization with config-server.
//
// DATA FLOW:
// 1. Input: TCP byte streams from config-server containing framed protobuf messages.
// 2. Logic: Uses safe-socket client socket; deserializes payloads through ConfigProtoHandler.
// 3. Output: Dispatches configuration updates to local atomic state and sends client update requests.
//
// KEY PARAMETERS:
// - addr: Target config-server endpoint string (host:port).
// =============================================================================

import (
	"fmt"
	"sync"
	"time"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	pb "github.com/Bastien-Antigravity/distributed-config/src/schemas"

	safesocket "github.com/Bastien-Antigravity/safe-socket"
)

// -----------------------------------------------------------------------------

// Client provides an interface to interact with the Config Server.
type Client struct {
	addr      string
	sock      safesocket.Socket
	Handler   *ConfigProtoHandler
	quit      chan struct{}
	closeOnce sync.Once
	backoff   *Backoff
	mu        sync.RWMutex
	watching  bool
	syncMu    sync.Mutex
}

// -----------------------------------------------------------------------------

// NewClient creates a new Config Client and connects to the server.
func NewClient(addr string, config *core.Config) (*Client, error) {
	h := NewConfigHandler("ClientHandler", config)

	c := &Client{
		addr:    addr,
		Handler: h,
		quit:    make(chan struct{}),
		backoff: NewBackoff(config),
	}

	h.SetOnGapDetected(func(expected, received uint64) {
		config.Logger.Warning("Client: Sequence gap detected (expected %d, got %d). Triggering FullRefresh...", expected, received)
		go func() {
			_, _ = c.FullRefresh()
		}()
	})

	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

// -----------------------------------------------------------------------------

// connect establishes the connection and acts the handshake.
func (c *Client) connect() error {
	// 1. Determine Identity from Config
	identity := c.Handler.parentConfig.Common.Name
	if identity == "" {
		identity = "distributed-config-client"
	}

	// 2. Build Profile String (syntax: profile:identity)
	profile := fmt.Sprintf("tcp-hello:%s", identity)

	// 3. Determine Public IP for binding
	publicIP := c.Handler.parentConfig.Common.PublicIP
	if publicIP == "" {
		publicIP = "127.0.0.1"
	}

	client, err := safesocket.CreateWithConfig(profile, c.addr, safesocket.SocketConfig{
		PublicIP: publicIP,
		Deadline: 30 * time.Second,
	}, "client", false)
	if err != nil {
		c.Handler.parentConfig.Logger.Error("Mock: Failed to create socket to %s (using safe-socket)", c.addr)
		return err
	}

	if err := client.Open(); err != nil {
		_ = client.Close()
		return err
	}

	c.mu.Lock()
	c.sock = client
	c.mu.Unlock()
	return nil
}

// -----------------------------------------------------------------------------

// Close closes the connection and stops the background listener.
func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		close(c.quit)
	})
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sock != nil {
		err := c.sock.Close()
		c.sock = nil
		return err
	}
	return nil
}

// -----------------------------------------------------------------------------

// Watch starts a background goroutine to handle asynchronous updates (BROADCASTs).
func (c *Client) Watch() {
	c.mu.Lock()
	if c.watching {
		c.mu.Unlock()
		return
	}
	c.watching = true
	c.mu.Unlock()

	go func() {
		defer func() {
			c.mu.Lock()
			c.watching = false
			c.mu.Unlock()
		}()
		attempt := 0
		for {
			select {
			case <-c.quit:
				return
			default:
				c.mu.RLock()
				sock := c.sock
				c.mu.RUnlock()

				if sock == nil {
					// Try to reconnect
					delay := c.backoff.GetDelay(attempt)
					c.Handler.parentConfig.Logger.Info("Client: Connection lost. Retrying in %v...", delay)
					select {
					case <-time.After(delay):
					case <-c.quit:
						return
					}
					if err := c.connect(); err == nil {
						c.Handler.parentConfig.Logger.Info("Client: Reconnected to %s", c.addr)
						attempt = 0
						// Re-sync after reconnection (send GET_SYNC; the main Watch loop will receive and process it)
						c.mu.RLock()
						reconnectedSock := c.sock
						c.mu.RUnlock()
						if reconnectedSock != nil {
							if syncData, err := c.Handler.HandleOutgoing(pb.ConfigMsg_GET_SYNC, nil); err == nil {
								_ = reconnectedSock.Send(syncData)
							}
						}
					} else {
						attempt++
					}
					continue
				}

				data, err := sock.Receive()
				if err != nil {
					// Connection likely closed or failed
					c.mu.Lock()
					if c.sock == sock {
						_ = c.sock.Close()
						c.sock = nil
					}
					c.mu.Unlock()
					continue
				}
				if len(data) > 0 {
					_ = c.Handler.HandleIncoming(data)
				}
			}
		}
	}()
}

// -----------------------------------------------------------------------------

// GetConfig fetches configuration from the server.
func (c *Client) GetConfig() (*core.Config, error) {
	return c.requestSync(pb.ConfigMsg_GET_SYNC)
}

// FullRefresh explicitly requests the full configuration from the server.
func (c *Client) FullRefresh() (*core.Config, error) {
	return c.requestSync(pb.ConfigMsg_FULL_REFRESH)
}

func (c *Client) requestSync(cmd pb.ConfigMsg_Cmd) (*core.Config, error) {
	// Send request via Handler
	data, err := c.Handler.HandleOutgoing(cmd, nil)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	sock := c.sock
	watching := c.watching
	c.mu.RUnlock()

	if sock != nil {
		c.syncMu.Lock()
		defer c.syncMu.Unlock()

		if watching {
			// Watch() loop is active and is the sole reader on sock.
			// Register a sync notification channel and wait with timeout.
			syncDone := make(chan struct{}, 1)
			c.Handler.SetOnSyncReceived(func() {
				select {
				case syncDone <- struct{}{}:
				default:
				}
			})
			defer c.Handler.SetOnSyncReceived(nil)

			if err := sock.Send(data); err != nil {
				return nil, err
			}

			select {
			case <-syncDone:
				return c.Handler.parentConfig, nil
			case <-time.After(5 * time.Second):
				return c.Handler.parentConfig, fmt.Errorf("timeout waiting for sync response from %s", c.addr)
			case <-c.quit:
				return nil, fmt.Errorf("client closed during sync")
			}
		} else {
			// Boot-time initial sync (Watch has not started yet): loop until GET_SYNC/FULL_REFRESH reply is received,
			// so interleaved BROADCAST_REGISTRY or BROADCAST_SYNC frames do not cause premature return.
			syncReceived := false
			c.Handler.SetOnSyncReceived(func() {
				syncReceived = true
			})
			defer c.Handler.SetOnSyncReceived(nil)

			if err := sock.Send(data); err != nil {
				return nil, err
			}

			for !syncReceived {
				respData, err := sock.Receive()
				if err != nil {
					return nil, err
				}
				if err := c.Handler.HandleIncoming(respData); err != nil {
					return nil, err
				}
			}
		}
	} else {
		// Mock behavior
		c.Handler.parentConfig.Logger.Info("Mock: Client.requestSync(%v) simulated", cmd)
	}

	return c.Handler.parentConfig, nil
}

// -----------------------------------------------------------------------------

// GetVersion returns the last seen server configuration version.
func (c *Client) GetVersion() uint64 {
	if c.Handler != nil {
		return c.Handler.GetLastSeenVersion()
	}
	return 0
}

// -----------------------------------------------------------------------------

// UpdateConfig sends the entire current live configuration to the server.
func (c *Client) UpdateConfig(cfg *core.Config) error {
	return c.UpdateConfigMap(cfg.LiveConfig.Load())
}

// UpdateConfigMap sends a specific configuration map to the server.
func (c *Client) UpdateConfigMap(m *map[string]map[string]string) error {
	if m == nil {
		return nil
	}
	data, err := c.Handler.HandleOutgoing(pb.ConfigMsg_PUT_SYNC, m)
	if err != nil {
		return err
	}

	c.mu.RLock()
	sock := c.sock
	c.mu.RUnlock()

	if sock != nil {
		return sock.Send(data)
	}
	return nil
}

// UpdateConfigMapWithOCC sends updates to the server with an expected version for Optimistic Concurrency Control.
// Returns ErrVersionConflict if the server version does not match expectedVersion.
func (c *Client) UpdateConfigMapWithOCC(updates interface{}, expectedVersion uint64) error {
	var payloadMap map[string]map[string]string
	var err error

	switch v := updates.(type) {
	case *map[string]map[string]string:
		if v != nil {
			payloadMap = *v
		}
	case map[string]map[string]string:
		payloadMap = v
	default:
		payloadMap, err = core.ParseSharePayload(updates)
		if err != nil {
			return fmt.Errorf("invalid OCC updates payload: %w", err)
		}
	}

	data, err := c.Handler.HandleOutgoingWithVersion(pb.ConfigMsg_PUT_SYNC, payloadMap, expectedVersion)
	if err != nil {
		return err
	}

	return c.sendWithUpdateAck(data)
}

// -----------------------------------------------------------------------------

// DeleteKeys sends a deletion request for specific section keys to the server.
func (c *Client) DeleteKeys(deletions map[string][]string) error {
	if len(deletions) == 0 {
		return nil
	}
	data, err := c.Handler.HandleOutgoing(pb.ConfigMsg_DELETE_SYNC, deletions)
	if err != nil {
		return err
	}

	c.mu.RLock()
	sock := c.sock
	c.mu.RUnlock()

	if sock != nil {
		return sock.Send(data)
	}
	return nil
}

// DeleteKeysWithOCC sends a deletion request with expected version for OCC validation.
// Returns ErrVersionConflict if the server version does not match expectedVersion.
func (c *Client) DeleteKeysWithOCC(deletions map[string][]string, expectedVersion uint64) error {
	if len(deletions) == 0 {
		return nil
	}
	data, err := c.Handler.HandleOutgoingWithVersion(pb.ConfigMsg_DELETE_SYNC, deletions, expectedVersion)
	if err != nil {
		return err
	}

	return c.sendWithUpdateAck(data)
}

// sendWithUpdateAck transmits a payload and waits for the server ACK or ERROR response.
func (c *Client) sendWithUpdateAck(data []byte) error {
	c.mu.RLock()
	sock := c.sock
	watching := c.watching
	c.mu.RUnlock()

	if sock == nil {
		c.Handler.parentConfig.Logger.Info("Mock: sendWithUpdateAck simulated")
		return nil
	}

	c.syncMu.Lock()
	defer c.syncMu.Unlock()

	if watching {
		resChan := make(chan error, 1)
		c.Handler.SetOnUpdateResult(func(ver uint64, err error) {
			select {
			case resChan <- err:
			default:
			}
		})
		defer c.Handler.SetOnUpdateResult(nil)

		if err := sock.Send(data); err != nil {
			return err
		}

		select {
		case err := <-resChan:
			return err
		case <-time.After(5 * time.Second):
			return fmt.Errorf("timeout waiting for server response from %s", c.addr)
		case <-c.quit:
			return fmt.Errorf("client closed during update")
		}
	} else {
		resReceived := false
		var resErr error
		c.Handler.SetOnUpdateResult(func(ver uint64, err error) {
			resReceived = true
			resErr = err
		})
		defer c.Handler.SetOnUpdateResult(nil)

		if err := sock.Send(data); err != nil {
			return err
		}

		for !resReceived {
			respData, err := sock.Receive()
			if err != nil {
				return err
			}
			_ = c.Handler.HandleIncoming(respData)
		}
		return resErr
	}
}

// IsConnected returns true if the client is currently connected.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sock != nil
}
