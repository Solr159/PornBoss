package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/jav"
)

func patchJavMetadata(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Title       *string `json:"title"`
		TitleEn     *string `json:"title_en"`
		StudioID    *int64  `json:"studio_id"`
		StudioName  string  `json:"studio_name"`
		ClearStudio bool    `json:"clear_studio"`
		SeriesID    *int64  `json:"series_id"`
		SeriesName  string  `json:"series_name"`
		ClearSeries bool    `json:"clear_series"`
		TagIDs      []int64 `json:"tag_ids"`
		Lock        *bool   `json:"lock"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}

	isEnglish := jav.CurrentMetadataLanguageIsEnglish()
	if lang := c.Query("metadata_lang"); lang == "en" {
		isEnglish = true
	} else if lang == "ja" || lang == "zh" {
		isEnglish = false
	}

	patch := dbpkg.JavMetadataPatch{
		Title:       req.Title,
		TitleEn:     req.TitleEn,
		StudioID:    req.StudioID,
		StudioName:  req.StudioName,
		ClearStudio: req.ClearStudio,
		SeriesID:    req.SeriesID,
		SeriesName:  req.SeriesName,
		ClearSeries: req.ClearSeries,
		TagIDs:      req.TagIDs,
		Lock:        req.Lock,
	}

	item, err := dbpkg.PatchJavMetadata(c.Request.Context(), id, patch, isEnglish)
	if err != nil {
		logging.Error("patch jav metadata id=%d: %v", id, err)
		if err.Error() == "jav not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, item)
}
