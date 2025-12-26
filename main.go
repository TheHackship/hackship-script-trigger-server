package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"script_trigger_server/runner"
	"script_trigger_server/server"
	"sync"
	"syscall"
)

func configFileParser(filePath string) (map[string]map[string]string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	serviceDir := make(map[string]map[string]string)

	if err := json.Unmarshal(data, &serviceDir); err != nil {
		return nil, err
	}

	return serviceDir, nil
}

func main() {
	// Buffered channel
	scriptChan := make(chan string, 10)

	// Parse config file
	configMap, err := configFileParser("./config.json")
	if err != nil {
		log.Fatalf("[MAIN]> failed to load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	var wg sync.WaitGroup

	// Start script runner
	wg.Add(1)
	r := runner.NewRunner(ctx, scriptChan)
	go func() {
		defer wg.Done()
		r.Start()
	}()

	// Start HTTP server in goroutine
	wg.Add(1)
	s := server.NewServer(ctx, configMap, scriptChan)
	go func() {
		defer wg.Done()
		s.Start()
	}()

	// Wait for shutdown signal
	<-ctx.Done()

	// Close channel to stop runner cleanly
	close(scriptChan)

	wg.Wait()
	log.Println("[MAIN]> Application exited")
}
