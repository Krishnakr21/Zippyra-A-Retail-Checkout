package local_status_server

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/zippyra/zippyra-connector/internal/erp_adapter"
	"github.com/zippyra/zippyra-connector/internal/logging"
)

type StatusMetrics struct {
	mu                   sync.RWMutex
	ErpType              string    `json:"erp_type"`
	LastPollAt           time.Time `json:"last_poll_at"`
	LastSuccessfulPushAt time.Time `json:"last_successful_push_at"`
	PendingJobsCount     int       `json:"pending_jobs_count"`
	LastError            string    `json:"last_error"`
}

type Server struct {
	port       int
	erpAdapter erp_adapter.ErpAdapter
	metrics    *StatusMetrics
	logger     *logging.Logger
	server     *http.Server
}

func NewServer(port int, erpType string, erpAdapter erp_adapter.ErpAdapter, logger *logging.Logger) (*Server, *StatusMetrics) {
	metrics := &StatusMetrics{
		ErpType: erpType,
	}

	s := &Server{
		port:       port,
		erpAdapter: erpAdapter,
		metrics:    metrics,
		logger:     logger,
	}

	return s, metrics
}

func (m *StatusMetrics) UpdatePoll(pendingCount int, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.LastPollAt = time.Now()
	m.PendingJobsCount = pendingCount
	if err != nil {
		m.LastError = err.Error()
	} else {
		m.LastError = ""
	}
}

func (m *StatusMetrics) UpdatePush(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err != nil {
		m.LastError = err.Error()
	} else {
		m.LastSuccessfulPushAt = time.Now()
		m.LastError = ""
	}
}

func (s *Server) Start() error {
	addr := fmt.Sprintf("127.0.0.1:%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to bind local status server to %s: %w", addr, err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/status", s.handleStatus)

	s.server = &http.Server{
		Handler: mux,
	}

	s.logger.Info("[StatusServer] Local status server listening on http://%s/status", addr)

	go func() {
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			s.logger.Error("[StatusServer] Server error: %v", err)
		}
	}()

	return nil
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.metrics.mu.RLock()
	erpType := s.metrics.ErpType
	lastPollAt := s.metrics.LastPollAt
	lastPushAt := s.metrics.LastSuccessfulPushAt
	pendingCount := s.metrics.PendingJobsCount
	lastErr := s.metrics.LastError
	s.metrics.mu.RUnlock()

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	erpHealth := "OK"
	if s.erpAdapter != nil {
		if err := s.erpAdapter.HealthCheck(ctx); err != nil {
			erpHealth = "ERROR: " + err.Error()
		}
	}

	status := "UP"
	if lastErr != "" || erpHealth != "OK" {
		status = "DEGRADED"
	}

	respMap := map[string]interface{}{
		"status":                  status,
		"erp_type":                erpType,
		"last_poll_at":            lastPollAt.Format(time.RFC3339),
		"last_successful_push_at": lastPushAt.Format(time.RFC3339),
		"pending_jobs_count":      pendingCount,
		"last_error":              lastErr,
		"erp_health":              erpHealth,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(respMap)
}

func (s *Server) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
