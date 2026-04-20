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
}

// -----------------------------------------------------------------------------

// NewClient creates a new Config Client and connects to the server.
func NewClient(addr string, config *core.Config) (*Client, error) {
	h := NewConfigHandler("ClientHandler", config)

	c := &Client{
		addr:    addr,
		Handler: h,
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

// Close closes the connection.
func (c *Client) Close() error {
	if c.sock != nil {
		return c.sock.Close()
	}
	return nil
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

	return &core.Config{}, nil
}

// -----------------------------------------------------------------------------

// UpdateConfig sends configuration updates to the server
func (c *Client) UpdateConfig(cfg *core.Config) error {
	data, err := c.Handler.HandleOutgoing(pb.ConfigMsg_PUT_SYNC, cfg.MemConfig)
	if err != nil {
		return err
	}
	if c.sock != nil {
		return c.sock.Send(data)
	}
	return nil
}
