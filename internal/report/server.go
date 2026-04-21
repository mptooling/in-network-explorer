package report

import (
	"context"
	"log/slog"
	"net"
	"net/http"

	explorer "github.com/pavlomaksymov/in-network-explorer/internal"
)

// PreviewServer serves an HTML report preview over HTTP for local development.
type PreviewServer struct {
	report *explorer.ProspectReport
	log    *slog.Logger
	addr   string
}

// NewPreviewServer creates a preview server that renders the given report on addr.
func NewPreviewServer(r *explorer.ProspectReport, addr string, log *slog.Logger) *PreviewServer {
	return &PreviewServer{report: r, addr: addr, log: log}
}

// NewPreviewHandler returns an http.Handler that serves HTML at / and JSON at /json.
func NewPreviewHandler(r *explorer.ProspectReport, log *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /json", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := RenderJSON(w, r); err != nil {
			log.Error("render json failed", "err", err)
			http.Error(w, "render error", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := RenderHTML(w, r); err != nil {
			log.Error("render html failed", "err", err)
			http.Error(w, "render error", http.StatusInternalServerError)
		}
	})

	return mux
}

// ListenAndServe starts the HTTP server. It blocks until ctx is cancelled.
func (s *PreviewServer) ListenAndServe(ctx context.Context) error {
	handler := NewPreviewHandler(s.report, s.log)

	srv := &http.Server{
		Addr:        s.addr,
		Handler:     handler,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	go func() {
		<-ctx.Done()
		_ = srv.Close()
	}()

	s.log.Info("preview server started", "addr", s.addr, "url", "http://"+s.addr)
	return srv.ListenAndServe()
}
