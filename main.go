package main

import (
    "flag"
    "fmt"
    "log"

	"github.com/goblinsan/mc-server-wrapper/config"
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
    // ...rest of your app logic...
}
