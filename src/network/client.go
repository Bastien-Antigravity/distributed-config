package network

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
	addr    string
	sock    safesocket.Socket
	Handler *ConfigProtoHandler
	quit    chan struct{}
	backoff *Backoff
	mu      sync.RWMutex
}

// -----------------------------------------------------------------------------

// NewClient creates a new Config Client and connects to the server.
func NewClient(addr string, config *core.Config) (*Client, error) {
	h := NewConfigHandler("ClientHandler", config)

	c := &Client{
		addr:    addr,
		Handler: h,
		quit:    make(chan struct{}),
		backoff: NewBackoff(),
	}
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

	client, err := safesocket.Create(profile, c.addr, "127.0.0.1", "client", false)
	if err != nil {
		c.Handler.parentConfig.Logger.Error("Mock: Failed to create socket to %s (using safe-socket)", c.addr)
		return err
	}

	if err := client.Open(); err != nil {
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
	close(c.quit)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sock != nil {
		return c.sock.Close()
	}
	return nil
}

// -----------------------------------------------------------------------------

// Watch starts a background goroutine to handle asynchronous updates (BROADCASTs).
func (c *Client) Watch() {
	go func() {
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
					time.Sleep(delay)
					if err := c.connect(); err == nil {
						c.Handler.parentConfig.Logger.Info("Client: Reconnected to %s", c.addr)
						attempt = 0
						// Re-sync after reconnection
						_, _ = c.GetConfig()
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
	// Send request via Handler
	data, err := c.Handler.HandleOutgoing(pb.ConfigMsg_GET_SYNC, nil)
	if err != nil {
		return nil, err
	}

	c.mu.RLock()
	sock := c.sock
	c.mu.RUnlock()

	if sock != nil {
		if err := sock.Send(data); err != nil {
			return nil, err
		}

		// Receive response (safe-socket handles framing)
		data, err := sock.Receive()
		if err != nil {
			return nil, err
		}

		// Pass actual read bytes
		if err := c.Handler.HandleIncoming(data); err != nil {
			return nil, err
		}
	} else {
		// Mock behavior
		c.Handler.parentConfig.Logger.Info("Mock: Client.GetConfig() simulated")
	}

	return c.Handler.parentConfig, nil
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

// IsConnected returns true if the client is currently connected.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.sock != nil
}
