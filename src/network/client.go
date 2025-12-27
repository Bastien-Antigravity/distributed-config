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
	sock    *safesocket.Socket
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
	// TODO: Determine correct Client IP dynamically if needed.
	client, err := safesocket.Create("tcp-hello", c.addr, "127.0.0.1")
	if err != nil {
		fmt.Printf("Mock: Failed to connect to %s (using safe-socket)\n", c.addr)
		return err
	}
	c.sock = client
	return nil
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
	data, err := c.Handler.HandleOutgoing(pb.ConfigMsg_get_mem_config, nil)
	if err != nil {
		return nil, err
	}

	if c.sock != nil {
		if err := c.sock.Send(data); err != nil {
			return nil, err
		}

		// Receive response (safe-socket handles framing)
		buf := make([]byte, 65535)
		n, err := c.sock.Receive(buf)
		if err != nil {
			return nil, err
		}

		// Pass actual read bytes
		if err := c.Handler.HandleIncoming(buf[:n]); err != nil {
			return nil, err
		}
	} else {
		// Mock behavior
		fmt.Println("Mock: Client.GetConfig() simulated")
	}

	return &core.Config{}, nil
}

// -----------------------------------------------------------------------------

// UpdateConfig sends configuration updates to the server
func (c *Client) UpdateConfig(cfg *core.Config) error {
	data, err := c.Handler.HandleOutgoing(pb.ConfigMsg_update_mem_config, nil)
	if err != nil {
		return err
	}
	if c.sock != nil {
		return c.sock.Send(data)
	}
	return nil
}
