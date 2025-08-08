package main

import (
	"log"
	"os"
)

var configDir string

func initializeConfig()  {

	// Create config directory
	// TODO - definitely need some error checking and alternate paths here
	configDir = os.Getenv("HOME")+"/.config/table-wrangler/"
	err := os.MkdirAll(configDir, 0755)
	if (err != nil) {
		log.Fatalf("Could not create/verify the config directory: %v", err)
	}
}
