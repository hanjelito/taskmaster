package web

import (
	"fmt"
	"net/http"
	"path/filepath"
	"taskmaster/internal/logger"
	"taskmaster/internal/process"
)

type Server struct {
	hub     *Hub
	manager *process.Manager
	logger  *logger.Logger
	port    int
}

func NewServer(port int, manager *process.Manager, logger *logger.Logger) *Server {
	hub := NewHub(logger)
	return &Server{
		hub:     hub,
		manager: manager,
		logger:  logger,
		port:    port,
	}
}

func (s *Server) Start() error {
	go s.hub.Run()

	http.HandleFunc("/", s.serveHome)
	http.HandleFunc("/ws", s.hub.ServeWS)
	http.HandleFunc("/api/status", s.handleStatus)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static/"))))

	addr := fmt.Sprintf(":%d", s.port)
	s.logger.Info("Starting web server on %s", addr)
	
	return http.ListenAndServe(addr, nil)
}

func (s *Server) serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	http.ServeFile(w, r, filepath.Join("web", "static", "index.html"))
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	
	status := s.manager.GetStatus()
	s.hub.BroadcastStatus(status)
	
	w.Write([]byte(`{"status": "ok"}`))
}

func (s *Server) GetHub() *Hub {
	return s.hub
}