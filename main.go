package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/goblinsan/mc-server-wrapper/config"
	"github.com/goblinsan/mc-server-wrapper/updater"
)

func main() {
	// Allow config file path as a command-line argument
	configPath := flag.String("config", "config.json", "Path to config file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Loaded config: %+v\n", cfg)

	// Example: get current version from file (if exists)
	currentVersion := ""
	if data, err := os.ReadFile(cfg.LastVersionFile); err == nil {
		currentVersion = string(data)
	}

	// Call the update logic, passing the download page URL from config
	updated, err := updater.UpdateServerIfNew(currentVersion, cfg.DownloadURL, cfg, updater.DefaultSymlinkUpdater)
	if err != nil {
		log.Fatalf("Update failed: %v", err)
	}
	if updated {
		fmt.Println("Update completed successfully.")
		// Optionally, write the new version to LastVersionFile here
	} else {
		fmt.Println("No update needed.")
	}
}
