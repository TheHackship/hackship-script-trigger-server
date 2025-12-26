package server

import (
	"encoding/json"
	"log"
	"net/http"
)

const AUTH_TOKEN = "HACKSHIP-COMM"

type RequestBody struct {
	Service string `json:"service"`
	Action  string `json:"action"`
}

type Server struct {
	serviceActionToScriptPathMap map[string]map[string]string
	scriptRunnerChan             chan string
}

func NewServer(mp map[string]map[string]string, runnerChan chan string) Server {
	return Server{
		serviceActionToScriptPathMap: mp,
		scriptRunnerChan:             runnerChan,
	}
}

func (s *Server) Start() {
	http.HandleFunc("/", s.requestHandler)
	log.Println("Server running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("HTTP Server Error:", err)
	}
}

func (s *Server) requestHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1MB

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

	serviceMap, serviceExists := s.serviceActionToScriptPathMap[payload.Service]
	if !serviceExists {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	scriptPath, scriptExists := serviceMap[payload.Action]
	if !scriptExists {
		http.Error(w, "action not found", http.StatusNotFound)
		return
	}

	s.scriptRunnerChan <- scriptPath

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("script started"))
}
