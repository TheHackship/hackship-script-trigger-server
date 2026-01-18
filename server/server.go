package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var AUTH_TOKEN string
var PORT int

func init() {
	AUTH_TOKEN = os.Getenv("AUTH_TOKEN")
	if AUTH_TOKEN == "" {
		AUTH_TOKEN = "HACKSHIP-COMM"
	}

	portStr := os.Getenv("PORT")
	if portStr == "" {
		PORT = 8080 // default
	} else {
		p, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatalf("[SERVER]> invalid PORT value: %s", portStr)
		}
		PORT = p
	}
}

type RequestBody struct {
	Service string `json:"service"`
	Action  string `json:"action"`
}

type Server struct {
	serviceActionToScriptPathMap map[string]map[string]string
	scriptRunnerChan             chan string
	ctx                          context.Context
	httpServer                   *http.Server
}

func NewServer(ctx context.Context, mp map[string]map[string]string, runnerChan chan string) *Server {
	mux := http.NewServeMux()

	s := &Server{
		serviceActionToScriptPathMap: mp,
		scriptRunnerChan:             runnerChan,
		ctx:                          ctx,
	}

	mux.HandleFunc("/", s.requestHandler)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", PORT),
		Handler: mux,
	}

	return s
}

func (s *Server) Start() {
	// Start HTTP server
	go func() {
		log.Printf("[SERVER]> Server running on http://localhost:%d\n", PORT)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[SERVER]> HTTP Server Error: %v", err)
		}
	}()

	// Wait for context cancellation
	<-s.ctx.Done()
	log.Println("[SERVER]> Shutting down HTTP server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("[SERVER]> HTTP shutdown error: %v", err)
	}
}

func (s *Server) requestHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Header.Get("Authorization") != AUTH_TOKEN {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var payload RequestBody
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	serviceMap, ok := s.serviceActionToScriptPathMap[payload.Service]
	if !ok {
		http.Error(w, "service not found", http.StatusNotFound)
		return
	}

	scriptPath, ok := serviceMap[payload.Action]
	if !ok {
		http.Error(w, "action not found", http.StatusNotFound)
		return
	}

	select {
	case s.scriptRunnerChan <- scriptPath:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("script started"))
	case <-ctx.Done():
		http.Error(w, "request cancelled", http.StatusRequestTimeout)
	}

}
