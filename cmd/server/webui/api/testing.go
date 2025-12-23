package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lich0821/ccNexus/internal/logger"
)

// testEndpoint tests an endpoint's connectivity
func (h *Handler) testEndpoint(w http.ResponseWriter, r *http.Request, name string) {
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Map endpoint name -> index in config (desktop semantics)
	endpoints := h.config.GetEndpoints()
	index := -1
	for i, ep := range endpoints {
		if ep.Name == name {
			index = i
			break
		}
	}
	if index < 0 {
		WriteError(w, http.StatusNotFound, "Endpoint not found")
		return
	}

	start := time.Now()
	resultJSON := h.endpoint.TestEndpointLight(index)
	latency := time.Since(start).Milliseconds()

	var parsed struct {
		Success bool   `json:"success"`
		Status  string `json:"status"`
		Method  string `json:"method"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal([]byte(resultJSON), &parsed); err != nil {
		logger.Error("Failed to parse endpoint test result: %v", err)
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"success": false,
			"latency": latency,
			"error":   "Invalid test result",
		})
		return
	}

	if parsed.Success {
		WriteJSON(w, http.StatusOK, map[string]interface{}{
			"success":  true,
			"latency":  latency,
			"response": parsed.Message,
			"status":   parsed.Status,
			"method":   parsed.Method,
		})
		return
	}

	WriteJSON(w, http.StatusOK, map[string]interface{}{
		"success": false,
		"latency": latency,
		"error":   parsed.Message,
		"status":  parsed.Status,
		"method":  parsed.Method,
	})
}

// handleFetchModels fetches available models from a provider
func (h *Handler) handleFetchModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		APIUrl      string `json:"apiUrl"`
		APIKey      string `json:"apiKey"`
		Transformer string `json:"transformer"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resultJSON := h.endpoint.FetchModels(req.APIUrl, req.APIKey, req.Transformer)
	var parsed struct {
		Success bool     `json:"success"`
		Message string   `json:"message"`
		Models  []string `json:"models"`
	}
	if err := json.Unmarshal([]byte(resultJSON), &parsed); err != nil {
		logger.Error("Failed to parse fetch-models result: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to fetch models")
		return
	}
	if !parsed.Success {
		WriteError(w, http.StatusBadRequest, parsed.Message)
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"models":  parsed.Models,
		"message": parsed.Message,
	})
}
