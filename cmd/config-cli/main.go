package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	distconf "github.com/Bastien-Antigravity/distributed-config"
)

func main() {
	// Parse command line flags
	profile := flag.String("profile", "standalone", "Configuration profile (standalone, test, preprod, production)")
	flag.Parse()

	fmt.Printf("Starting Config CLI with profile: %s\n", *profile)

	// Initialize Configuration
	config := distconf.New(*profile)

	// Print initial configuration
	printConfig(config)

	// Setup update listener
	config.OnMemConfUpdate(func(updates map[string]map[string]string) {
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
	if config.MemConfig == nil || len(config.MemConfig) == 0 {
		fmt.Println("  (Empty)")
		return
	}

	for section, kv := range config.MemConfig {
		fmt.Printf("  [%s]\n", section)
		for k, v := range kv {
			fmt.Printf("    %s = %s\n", k, v)
		}
	}
}
