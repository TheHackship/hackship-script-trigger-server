package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type RequestBody struct {
	Service string `json:"service"`
	Action  string `json:"action"`
}

const AUTH_TOKEN = "HACKSHIP-COMM"

// global config map
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

func scriptRuntime(scriptPath string) {
	go func() {
		cmd := exec.Command("/bin/bash", scriptPath)
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Printf("script failed: %v\noutput: %s", err, output)
			return
		}
		log.Printf("script executed successfully:\n%s", output)
	}()
}

func requestHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Auth check
	if r.Header.Get("Authorization") != AUTH_TOKEN {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload RequestBody
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	serviceMap, serviceExists := serviceDir[payload.Service]
	if !serviceExists {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	scriptPath, scriptExists := serviceMap[payload.Action]
	if !scriptExists {
		http.Error(w, "action not found", http.StatusNotFound)
		return
	}

	scriptRuntime(scriptPath)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("script started"))
}

func main() {
	rootDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	filePath := filepath.Join(rootDir, "config.json")

	if err := jsonParser(filePath); err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	http.HandleFunc("/", requestHandler)

	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("HTTP Server Error:", err)
	}
}
