package report

import (
	"fmt"
	"log/slog"
	"net/http"
)

// Server serves the prospect report as a live HTML page for local preview.
type Server struct {
	report *Report
	log    *slog.Logger
}

// NewServer creates a preview server for the given report.
func NewServer(r *Report, log *slog.Logger) *Server {
	return &Server{report: r, log: log}
}

// ListenAndServe starts the HTTP server on the given address (e.g. ":8080").
func (s *Server) ListenAndServe(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", s.handleHTML)
	mux.HandleFunc("GET /api/report", s.handleJSON)

	s.log.Info("preview server starting", "addr", addr)
	return http.ListenAndServe(addr, mux)
}

func (s *Server) handleHTML(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.report.WriteHTML(w); err != nil {
		s.log.Error("render html", "error", err)
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}

func (s *Server) handleJSON(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if err := s.report.WriteJSON(w); err != nil {
		s.log.Error("render json", "error", err)
		http.Error(w, "render failed", http.StatusInternalServerError)
	}
}

// Addr returns the full URL for the given listen address.
func Addr(listen string) string {
	return fmt.Sprintf("http://localhost%s", listen)
}
