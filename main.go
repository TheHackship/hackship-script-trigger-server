package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"script_trigger_server/runner"
	"script_trigger_server/server"
)

var scriptChan = make(chan string)
var serviceDir map[string]map[string]string

func jsonParser(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	serviceDir = make(map[string]map[string]string)

	if err := json.Unmarshal(data, &serviceDir); err != nil {
		return err
	}

	return nil
}

func main() {
	// Init serviceScriptPath map
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	filePath := filepath.Join(rootDir, "config.json")
	if err := jsonParser(filePath); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Start script runner
	r := runner.NewRunner(scriptChan)
	go r.Start()

	// Start HTTP server in goroutine
	s := server.NewServer(serviceDir, scriptChan)
	go s.Start()

	select {}
}
