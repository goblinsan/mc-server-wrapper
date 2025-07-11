package main

import (
	"fmt"
	"log"
	"os"

	"github.com/goblinsan/mc-server-wrapper/config"
)

func main() {
	configPath := "config.json" // or use a flag/env var for flexibility
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Loaded config: %+v\n", cfg)
	// ...rest of your app logic...
}
