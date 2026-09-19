package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

const (
	defaultPort    = "8080"
	defaultVersion = "v0.1.1"
	serviceName    = "notiflex-api"
)

var eventSequence uint64

type server struct {
	version string
	logger  *slog.Logger
}

type eventRequest struct {
	Type      string         `json:"type"`
	Recipient string         `json:"recipient"`
	Channel   string         `json:"channel,omitempty"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type eventResponse struct {
	Accepted bool   `json:"accepted"`
	EventID  string `json:"eventId"`
	Message  string `json:"message"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	port := envOrDefault("PORT", defaultPort)
	version := envOrDefault("APP_VERSION", defaultVersion)

	api := &server{
		version: version,
		logger:  logger,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", api.handleRoot)
	mux.HandleFunc("/healthz", api.handleHealthz)
	mux.HandleFunc("/readyz", api.handleReadyz)
	mux.HandleFunc("/v1/events", api.handleEvents)

	httpServer := &http.Server{
		Addr:              ":" + port,
		Handler:           requestLogger(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("starting server", "service", serviceName, "version", version, "port", port)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-shutdownCtx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	logger.Info("shutting down server")
	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
}

func (s *server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"service": serviceName,
		"version": s.version,
		"message": "Notiflex API is running",
	})
}

func (s *server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) handleReadyz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ready"})
}

func (s *server) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var event eventRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&event); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON request body")
		return
	}

	event.Channel = strings.ToLower(strings.TrimSpace(event.Channel))
	if event.Channel == "" {
		event.Channel = "email"
	}

	if err := validateEvent(event); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	eventID := newEventID()
	s.logger.Info(
		"event accepted",
		"eventId", eventID,
		"type", strings.TrimSpace(event.Type),
		"recipient", strings.TrimSpace(event.Recipient),
		"channel", event.Channel,
	)

	writeJSON(w, http.StatusAccepted, eventResponse{
		Accepted: true,
		EventID:  eventID,
		Message:  "event accepted",
	})
}

func validateEvent(event eventRequest) error {
	if strings.TrimSpace(event.Type) == "" {
		return errors.New("type is required")
	}
	if strings.TrimSpace(event.Recipient) == "" {
		return errors.New("recipient is required")
	}

	switch event.Channel {
	case "email", "sms", "push":
		return nil
	default:
		return errors.New("channel must be one of email, sms, push")
	}
}

func newEventID() string {
	sequence := atomic.AddUint64(&eventSequence, 1)
	random := make([]byte, 3)
	if _, err := rand.Read(random); err != nil {
		return fmt.Sprintf("evt_%s_%06d", time.Now().UTC().Format("20060102T150405Z"), sequence)
	}

	return fmt.Sprintf("evt_%s_%06d_%s", time.Now().UTC().Format("20060102T150405Z"), sequence, hex.EncodeToString(random))
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path == "/healthz" || r.URL.Path == "/readyz" {
			return
		}
		logger.Info("request completed", "method", r.Method, "path", r.URL.Path, "durationMs", time.Since(started).Milliseconds())
	})
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}

func envOrDefault(name, fallback string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	return value
}
