package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
)

func listVideoMarkers(c *gin.Context) {
	videoID, ok := parseVideoIDParam(c)
	if !ok {
		return
	}
	items, err := dbpkg.ListVideoMarkers(c.Request.Context(), videoID)
	if err != nil {
		logging.Error("list video markers: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func createVideoMarker(c *gin.Context) {
	videoID, ok := parseVideoIDParam(c)
	if !ok {
		return
	}
	var req struct {
		TimeSec float64 `json:"time_sec"`
		Note    string  `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	marker, err := dbpkg.CreateVideoMarker(c.Request.Context(), videoID, req.TimeSec, req.Note)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, marker)
}

func patchVideoMarker(c *gin.Context) {
	videoID, ok := parseVideoIDParam(c)
	if !ok {
		return
	}
	markerID, err := strconv.ParseInt(c.Param("markerId"), 10, 64)
	if err != nil || markerID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid marker id"})
		return
	}
	var req struct {
		TimeSec *float64 `json:"time_sec"`
		Note    *string  `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	marker, err := dbpkg.UpdateVideoMarker(c.Request.Context(), videoID, markerID, req.TimeSec, req.Note)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, marker)
}

func deleteVideoMarker(c *gin.Context) {
	videoID, ok := parseVideoIDParam(c)
	if !ok {
		return
	}
	markerID, err := strconv.ParseInt(c.Param("markerId"), 10, 64)
	if err != nil || markerID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid marker id"})
		return
	}
	if err := dbpkg.DeleteVideoMarker(c.Request.Context(), videoID, markerID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func parseVideoIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return 0, false
	}
	return id, true
}
