package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common"
	"pornboss/internal/manager"
)

// getJavCover serves a downloaded JAV cover if present; otherwise enqueues and returns 404.
func getJavCover(c *gin.Context) {
	code := c.Param("code")
	cfg := common.AppConfig
	if cfg == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "config not loaded"})
		return
	}

	if path, ok := manager.FindCoverPath(cfg.JavCoverDir, code); ok {
		c.File(path)
		return
	}

	if common.CoverManager != nil {
		common.CoverManager.Enqueue(code)
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "cover not found"})
}
