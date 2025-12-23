package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/lich0821/ccNexus/internal/config"
	"github.com/lich0821/ccNexus/internal/logger"
)

type webdavJSON map[string]interface{}

func parseWebDAVJSON(s string) (webdavJSON, error) {
	var out webdavJSON
	s = strings.TrimSpace(s)
	if s == "" {
		return webdavJSON{}, nil
	}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, err
	}
	return out, nil
}

// handleWebDAVConfig supports GET (current config) and PUT (update config)
func (h *Handler) handleWebDAVConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		cfg := h.config.GetWebDAV()
		if cfg == nil {
			WriteSuccess(w, map[string]interface{}{
				"configured": false,
				"url":        "",
				"username":   "",
				"hasPassword": false,
			})
			return
		}
		WriteSuccess(w, map[string]interface{}{
			"configured":  true,
			"url":         cfg.URL,
			"username":    cfg.Username,
			"hasPassword": cfg.Password != "",
		})
	case http.MethodPut:
		var req struct {
			URL      string `json:"url"`
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if strings.TrimSpace(req.URL) == "" {
			WriteError(w, http.StatusBadRequest, "url is required")
			return
		}
		if req.Password == "" {
			if existing := h.config.GetWebDAV(); existing != nil && existing.Password != "" {
				req.Password = existing.Password
			}
		}
		if err := h.webdav.UpdateWebDAVConfig(strings.TrimSpace(req.URL), strings.TrimSpace(req.Username), req.Password); err != nil {
			logger.Error("Failed to update WebDAV config: %v", err)
			WriteError(w, http.StatusInternalServerError, "Failed to update WebDAV config")
			return
		}
		WriteSuccess(w, map[string]interface{}{
			"message": "WebDAV configuration updated successfully",
		})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleWebDAVTest tests WebDAV connection with provided credentials (does not persist)
func (h *Handler) handleWebDAVTest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		URL      string `json:"url"`
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	resultStr := h.webdav.TestWebDAVConnection(strings.TrimSpace(req.URL), strings.TrimSpace(req.Username), req.Password)
	result, err := parseWebDAVJSON(resultStr)
	if err != nil {
		logger.Error("Failed to parse WebDAV test result: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to parse WebDAV test result")
		return
	}
	WriteSuccess(w, result)
}

// handleWebDAVBackups lists or deletes backups
func (h *Handler) handleWebDAVBackups(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		resultStr := h.webdav.ListWebDAVBackups()
		result, err := parseWebDAVJSON(resultStr)
		if err != nil {
			logger.Error("Failed to parse WebDAV backups list: %v", err)
			WriteError(w, http.StatusInternalServerError, "Failed to list WebDAV backups")
			return
		}
		WriteSuccess(w, result)
	case http.MethodDelete:
		var req struct {
			Filenames []string `json:"filenames"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			WriteError(w, http.StatusBadRequest, "Invalid request body")
			return
		}
		if len(req.Filenames) == 0 {
			WriteError(w, http.StatusBadRequest, "filenames is required")
			return
		}
		if err := h.webdav.DeleteWebDAVBackups(req.Filenames); err != nil {
			logger.Error("Failed to delete WebDAV backups: %v", err)
			WriteError(w, http.StatusInternalServerError, "Failed to delete WebDAV backups")
			return
		}
		WriteSuccess(w, map[string]interface{}{
			"message": "Backups deleted successfully",
		})
	default:
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleWebDAVBackup triggers a backup to WebDAV
func (h *Handler) handleWebDAVBackup(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Filename string `json:"filename"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	filename := strings.TrimSpace(req.Filename)
	if filename == "" {
		filename = "backup-" + time.Now().Format("20060102-150405") + ".db"
	}

	if err := h.webdav.BackupToWebDAV(filename); err != nil {
		logger.Error("WebDAV backup failed: %v", err)
		WriteError(w, http.StatusInternalServerError, "WebDAV backup failed")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"message":  "Backup created successfully",
		"filename": filename,
	})
}

// handleWebDAVRestore restores from a WebDAV backup
func (h *Handler) handleWebDAVRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Filename string `json:"filename"`
		Choice   string `json:"choice"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	filename := strings.TrimSpace(req.Filename)
	choice := strings.TrimSpace(req.Choice)
	if filename == "" {
		WriteError(w, http.StatusBadRequest, "filename is required")
		return
	}
	if choice == "" {
		choice = "local"
	}
	if choice != "remote" && choice != "local" && choice != "keep_local" {
		WriteError(w, http.StatusBadRequest, "choice must be one of: remote, local, keep_local")
		return
	}

	reload := func(cfg interface{}) error {
		// cfg is *config.Config at runtime; keep signature local to avoid import cycles here
		return nil
	}
	_ = reload

	// Use the same semantics as desktop: update proxy config after merge
	if err := h.webdav.RestoreFromWebDAV(filename, choice, func(cfg *config.Config) error {
		h.config = cfg
		return h.proxy.UpdateConfig(cfg)
	}); err != nil {
		logger.Error("WebDAV restore failed: %v", err)
		WriteError(w, http.StatusInternalServerError, "WebDAV restore failed")
		return
	}

	WriteSuccess(w, map[string]interface{}{
		"message": "Restore completed successfully",
	})
}

// handleWebDAVConflict checks conflicts for a given backup filename
func (h *Handler) handleWebDAVConflict(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	filename := strings.TrimSpace(r.URL.Query().Get("filename"))
	if filename == "" {
		WriteError(w, http.StatusBadRequest, "filename is required")
		return
	}

	resultStr := h.webdav.DetectWebDAVConflict(filename)
	result, err := parseWebDAVJSON(resultStr)
	if err != nil {
		logger.Error("Failed to parse WebDAV conflict result: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to detect conflicts")
		return
	}
	WriteSuccess(w, result)
}
