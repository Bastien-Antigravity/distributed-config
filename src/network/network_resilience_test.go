package network

import (
	"fmt"
	"testing"
	"time"

	"github.com/Bastien-Antigravity/distributed-config/src/core"
	"github.com/Bastien-Antigravity/distributed-config/src/utils"
	safesocket "github.com/Bastien-Antigravity/safe-socket"
)

func TestNetworkResilience_Reconnection(t *testing.T) {
	addr := "127.0.0.1:9999"

	server1, err := safesocket.Create("tcp-hello:server", addr, "127.0.0.1", "server", false)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Use a channel to control server lifecycle
	serverStop1 := make(chan struct{})
	serverReady1 := make(chan struct{})
	go func() {
		if err := server1.Listen(); err != nil {
			fmt.Printf("Server Listen Error: %v\n", err)
			return
		}
		close(serverReady1)

		for {
			select {
			case <-serverStop1:
				_ = server1.Close()
				return
			default:
				conn, err := server1.Accept()
				if err != nil {
					return
				}
				go func() {
					defer func() { _ = conn.Close() }()
					for {
						_, err := conn.ReadMessage()
						if err != nil {
							return
						}
					}
				}()
			}
		}
	}()

	// Wait for server to be ready
	select {
	case <-serverReady1:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatalf("Server failed to start in time")
	}

	// 2. Setup Client
	cfg := &core.Config{}
	cfg.Common.Name = "test-client"
	cfg.Logger = utils.EnsureSafeLogger(nil)
	client, err := NewClient(addr, cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer func() { _ = client.Close() }()

	client.Watch()
	time.Sleep(200 * time.Millisecond)

	// 3. Kill Server
	fmt.Println("Killing server...")
	close(serverStop1)
	time.Sleep(1 * time.Second)

	// 4. Restart Server
	fmt.Println("Restarting server...")
	serverStop2 := make(chan struct{})
	serverReady2 := make(chan struct{})
	newServer, err := safesocket.Create("tcp-hello:server", addr, "127.0.0.1", "server", false)
	if err != nil {
		t.Fatalf("Failed to create new server: %v", err)
	}

	go func() {
		if err := newServer.Listen(); err != nil {
			return
		}
		close(serverReady2)
		for {
			select {
			case <-serverStop2:
				_ = newServer.Close()
				return
			default:
				conn, err := newServer.Accept()
				if err != nil {
					return
				}
				go func() {
					defer func() { _ = conn.Close() }()
					for {
						_, err := conn.ReadMessage()
						if err != nil {
							return
						}
					}
				}()
			}
		}
	}()
	<-serverReady2
	defer close(serverStop2)

	// 5. Wait for Client to reconnect
	reconnected := false
	for i := 0; i < 20; i++ {
		if client.IsConnected() {
			reconnected = true
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if !reconnected {
		t.Errorf("Client failed to reconnect after server restart")
	}
}
