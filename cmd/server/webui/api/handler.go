package api

import (
	"net/http"

	"github.com/lich0821/ccNexus/internal/config"
	"github.com/lich0821/ccNexus/internal/proxy"
	"github.com/lich0821/ccNexus/internal/service"
	"github.com/lich0821/ccNexus/internal/storage"
)

// Handler handles API requests
type Handler struct {
	config  *config.Config
	proxy   *proxy.Proxy
	storage *storage.SQLiteStorage
	endpoint *service.EndpointService
	webdav  *service.WebDAVService
}

// NewHandler creates a new API handler
func NewHandler(cfg *config.Config, p *proxy.Proxy, s *storage.SQLiteStorage, version string) *Handler {
	return &Handler{
		config:  cfg,
		proxy:   p,
		storage: s,
		endpoint: service.NewEndpointService(cfg, p, s),
		webdav:  service.NewWebDAVService(cfg, s, version),
	}
}

// RegisterRoutes registers all API routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Endpoint management
	mux.HandleFunc("/api/endpoints", h.handleEndpoints)
	mux.HandleFunc("/api/endpoints/", h.handleEndpointByName)
	mux.HandleFunc("/api/endpoints/current", h.handleCurrentEndpoint)
	mux.HandleFunc("/api/endpoints/switch", h.handleSwitchEndpoint)
	mux.HandleFunc("/api/endpoints/reorder", h.handleReorderEndpoints)
	mux.HandleFunc("/api/endpoints/fetch-models", h.handleFetchModels)

	// Statistics
	mux.HandleFunc("/api/stats/summary", h.handleStatsSummary)
	mux.HandleFunc("/api/stats/daily", h.handleStatsDaily)
	mux.HandleFunc("/api/stats/weekly", h.handleStatsWeekly)
	mux.HandleFunc("/api/stats/monthly", h.handleStatsMonthly)
	mux.HandleFunc("/api/stats/trends", h.handleStatsTrends)

	// Configuration
	mux.HandleFunc("/api/config", h.handleConfig)
	mux.HandleFunc("/api/config/port", h.handleConfigPort)
	mux.HandleFunc("/api/config/log-level", h.handleConfigLogLevel)

	// Real-time events
	mux.HandleFunc("/api/events", h.handleEvents)

	// WebDAV
	mux.HandleFunc("/api/webdav/config", h.handleWebDAVConfig)
	mux.HandleFunc("/api/webdav/test", h.handleWebDAVTest)
	mux.HandleFunc("/api/webdav/backups", h.handleWebDAVBackups)
	mux.HandleFunc("/api/webdav/backup", h.handleWebDAVBackup)
	mux.HandleFunc("/api/webdav/restore", h.handleWebDAVRestore)
	mux.HandleFunc("/api/webdav/conflict", h.handleWebDAVConflict)
}
