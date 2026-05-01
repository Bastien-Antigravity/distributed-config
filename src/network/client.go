package network

import (
	"fmt"
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
}

// -----------------------------------------------------------------------------

// NewClient creates a new Config Client and connects to the server.
func NewClient(addr string, config *core.Config) (*Client, error) {
	h := NewConfigHandler("ClientHandler", config)

	c := &Client{
		addr:    addr,
		Handler: h,
		quit:    make(chan struct{}),
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
		c.Handler.parentConfig.Logger.Error("Mock: Failed to connect to %s (using safe-socket)", c.addr)
		return err
	}
	c.sock = client
	return c.sock.Open()
}

// -----------------------------------------------------------------------------

// Close closes the connection and stops the background listener.
func (c *Client) Close() error {
	close(c.quit)
	if c.sock != nil {
		return c.sock.Close()
	}
	return nil
}

// -----------------------------------------------------------------------------

// Watch starts a background goroutine to handle asynchronous updates (BROADCASTs).
func (c *Client) Watch() {
	go func() {
		for {
			select {
			case <-c.quit:
				return
			default:
				if c.sock == nil {
					return
				}
				data, err := c.sock.Receive()
				if err != nil {
					// Connection likely closed
					return
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

	if c.sock != nil {
		if err := c.sock.Send(data); err != nil {
			return nil, err
		}

		// Receive response (safe-socket handles framing)
		data, err := c.sock.Receive()
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
	if c.sock != nil {
		return c.sock.Send(data)
	}
	return nil
}
