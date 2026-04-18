package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/util"
)

const maxPageSize = 500

func getConfig(c *gin.Context) {
	cfg, err := dbpkg.ListConfig(c.Request.Context())
	if err != nil {
		logging.Error("list config error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, cfg)
}

func updateConfig(c *gin.Context) {
	type playerHotkeyPayload struct {
		Key    string  `json:"key"`
		Action string  `json:"action"`
		Amount float64 `json:"amount"`
	}

	var req struct {
		VideoPageSize *int                  `json:"video_page_size"`
		JavPageSize   *int                  `json:"jav_page_size"`
		IdolPageSize  *int                  `json:"idol_page_size"`
		VideoSort     string                `json:"video_sort"`
		JavSort       string                `json:"jav_sort"`
		IdolSort      string                `json:"idol_sort"`
		ProxyPort     *int                  `json:"proxy_port"`
		PlayerHotkeys []playerHotkeyPayload `json:"player_hotkeys"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	entries := make(map[string]string)
	clampSize := func(n int) (string, bool) {
		if n <= 0 {
			return "", false
		}
		if n > maxPageSize {
			n = maxPageSize
		}
		return strconv.Itoa(n), true
	}

	if req.VideoPageSize != nil {
		if v, ok := clampSize(*req.VideoPageSize); ok {
			entries["video_page_size"] = v
		}
	}
	if req.JavPageSize != nil {
		if v, ok := clampSize(*req.JavPageSize); ok {
			entries["jav_page_size"] = v
		}
	}
	if req.IdolPageSize != nil {
		if v, ok := clampSize(*req.IdolPageSize); ok {
			entries["idol_page_size"] = v
		}
	}
	if s := strings.ToLower(strings.TrimSpace(req.VideoSort)); s != "" {
		switch s {
		case "recent", "filename", "duration", "play_count":
			entries["video_sort"] = s
		default:
			// ignore invalid values
		}
	}
	if s := strings.ToLower(strings.TrimSpace(req.JavSort)); s != "" {
		switch s {
		case "recent", "code", "duration", "release", "play_count":
			entries["jav_sort"] = s
		default:
			// ignore invalid values
		}
	}
	if s := strings.ToLower(strings.TrimSpace(req.IdolSort)); s != "" {
		switch s {
		case "work", "birth", "height", "bust", "hips", "waist", "measurements", "cup":
			entries["idol_sort"] = s
		default:
			// ignore invalid values
		}
	}
	if req.ProxyPort != nil {
		port := *req.ProxyPort
		if port <= 0 {
			entries["proxy_port"] = ""
		} else if port <= 65535 {
			entries["proxy_port"] = strconv.Itoa(port)
		}
	}
	if req.PlayerHotkeys != nil {
		clean := make([]playerHotkeyPayload, 0, len(req.PlayerHotkeys))
		seen := make(map[string]struct{}, len(req.PlayerHotkeys))
		for _, item := range req.PlayerHotkeys {
			key := strings.TrimSpace(item.Key)
			action := strings.ToLower(strings.TrimSpace(item.Action))
			if len(key) == 1 {
				key = strings.ToLower(key)
			}
			if key == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "player hotkey key required"})
				return
			}
			if key == " " || strings.EqualFold(key, "spacebar") || strings.EqualFold(key, "escape") {
				c.JSON(http.StatusBadRequest, gin.H{"error": "space and escape are reserved"})
				return
			}
			if _, ok := seen[key]; ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "duplicate player hotkeys"})
				return
			}
			if action != "seek" && action != "volume" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid player hotkey action"})
				return
			}
			if item.Amount == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "player hotkey amount required"})
				return
			}
			if action == "volume" && (item.Amount < -100 || item.Amount > 100) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "player volume hotkey out of range"})
				return
			}
			seen[key] = struct{}{}
			clean = append(clean, playerHotkeyPayload{
				Key:    key,
				Action: action,
				Amount: item.Amount,
			})
		}
		raw, err := json.Marshal(clean)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		entries["player_hotkeys"] = string(raw)
	}

	if err := dbpkg.UpsertConfig(c.Request.Context(), entries); err != nil {
		logging.Error("update config error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	cfg, err := dbpkg.ListConfig(c.Request.Context())
	if err != nil {
		logging.Error("list config after update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	util.SetProxyPortFromString(cfg["proxy_port"])
	c.JSON(http.StatusOK, cfg)
}
