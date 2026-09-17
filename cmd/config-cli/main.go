package main

// =============================================================================
// ESSENTIAL PROCESS:
// Command-line interface for inspecting initial configuration state and
// subscribing to live dynamic updates from the fleet configuration server.
//
// DATA FLOW:
// 1. Input: CLI flag `--profile` (defaults to "standalone").
// 2. Logic: Instantiates distconf.New(profile), prints snapshot, listens for live updates.
// 3. Output: Formatted configuration key-values rendered to stdout in real-time.
//
// KEY PARAMETERS:
// - profile: Selected operational profile (standalone, test, staging, production).
// =============================================================================

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	distconf "github.com/Bastien-Antigravity/distributed-config"
)

// -----------------------------------------------------------------------------

func main() {
	// Parse command line flags
	profile := flag.String("profile", "standalone", "Configuration profile (standalone, test, staging, production)")
	flag.Parse()

	fmt.Printf("Starting Config CLI with profile: %s\n", *profile)

	// Initialize Configuration
	config := distconf.New(*profile)

	// Print initial configuration
	printConfig(config)

	// Setup update listener
	config.OnLiveConfUpdate(func(updates map[string]map[string]string) {
		fmt.Println("\n[Update Received] Configuration changed:")
		for section, kv := range updates {
			fmt.Printf("  [%s]\n", section)
			for k, v := range kv {
				fmt.Printf("    %s = %s\n", k, v)
			}
		}
	})

	// Keep alive to receive updates
	fmt.Println("\nListening for updates... (Press Ctrl+C to exit)")
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down...")
}

func printConfig(config *distconf.Config) {
	fmt.Println("\nCurrent Configuration:")

	livePtr := config.LiveConfig.Load()
	if livePtr == nil || len(*livePtr) == 0 {
		fmt.Println("  (Empty)")
		return
	}

	for section, kv := range *livePtr {
		fmt.Printf("  [%s]\n", section)
		for k, v := range kv {
			fmt.Printf("    %s = %s\n", k, v)
		}
	}
}
