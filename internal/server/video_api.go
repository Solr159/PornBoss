package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common"
	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/manager"
	"pornboss/internal/util"
)

func listVideos(c *gin.Context) {
	limit := queryInt(c, "limit", 100)
	offset := queryInt(c, "offset", 0)
	tagFilter := parseTagQuery(c.Query("tags"))
	search := strings.TrimSpace(c.Query("search"))
	sort := strings.TrimSpace(c.Query("sort"))
	seedParam := strings.TrimSpace(c.Query("seed"))
	var seed *int64
	if seedParam != "" {
		parsed, err := strconv.ParseInt(seedParam, 10, 64)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid seed"})
			return
		}
		seed = &parsed
	}

	videos, err := dbpkg.ListVideos(c.Request.Context(), limit, offset, tagFilter, search, sort, seed)
	if err != nil {
		logging.Error("list videos error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	total, err := dbpkg.CountVideos(c.Request.Context(), tagFilter, search)
	if err != nil {
		logging.Error("count videos error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": videos,
		"total": total,
	})
}

func getVideo(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	video, err := dbpkg.GetVideo(c.Request.Context(), id)
	if err != nil {
		logging.Error("get video error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if video == nil {
		c.Status(http.StatusNotFound)
		return
	}
	c.JSON(http.StatusOK, video)
}

func incrementVideoPlayCount(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := dbpkg.IncrementVideoPlayCount(c.Request.Context(), id); err != nil {
		logging.Error("increment play count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func streamVideo(c *gin.Context) {
	fullPath := ""
	if rawPath := strings.TrimSpace(c.Query("path")); rawPath != "" {
		var err error
		fullPath, err = resolveStreamPathFromQuery(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		_, resolvedPath, ok := loadVideoByID(c)
		if !ok {
			return
		}
		fullPath = resolvedPath
	}
	serveVideoFile(c, fullPath)
}

func resolveStreamPathFromQuery(c *gin.Context) (string, error) {
	rawPath := strings.TrimSpace(c.Query("path"))
	rawDirPath := strings.TrimSpace(c.Query("dir_path"))
	fullPath, _, err := resolveVideoPath(rawPath, rawDirPath)
	return fullPath, err
}

func serveVideoFile(c *gin.Context, fullPath string) {
	if _, err := os.Stat(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.Status(http.StatusNotFound)
			return
		}
		logging.Error("stat stream file error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.File(fullPath)
}

func openVideoFile(c *gin.Context) {
	fullPath, dirPath, err := resolveVideoPathFromBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}
	if err := util.OpenFile(fullPath); err != nil {
		logging.Error("open video file error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "open file failed"})
		return
	}
	incrementPlayCountByPath(c.Request.Context(), dirPath, fullPath)
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func revealVideoLocation(c *gin.Context) {
	fullPath, _, err := resolveVideoPathFromBody(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := ensureVideoFileExists(c, fullPath); err != nil {
		return
	}
	if err := util.RevealFile(fullPath); err != nil {
		logging.Error("reveal video file error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reveal file failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type videoPathRequest struct {
	Path    string `json:"path"`
	DirPath string `json:"dir_path"`
}

func resolveVideoPathFromBody(c *gin.Context) (string, string, error) {
	var req videoPathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		return "", "", errors.New("invalid payload")
	}
	fullPath, dirPath, err := resolveVideoPath(req.Path, req.DirPath)
	return fullPath, dirPath, err
}

func resolveVideoPath(rawPath, rawDirPath string) (string, string, error) {
	if strings.TrimSpace(rawPath) == "" || strings.TrimSpace(rawDirPath) == "" {
		return "", "", errors.New("path and dir_path are required")
	}

	dirPath := filepath.Clean(rawDirPath)
	if dirPath == "." || !filepath.IsAbs(dirPath) {
		return "", "", errors.New("invalid dir_path")
	}

	cleanPath := filepath.Clean(filepath.FromSlash(rawPath))
	if cleanPath == "." {
		return "", "", errors.New("invalid path")
	}

	fullPath := cleanPath
	if !filepath.IsAbs(cleanPath) {
		fullPath = filepath.Join(dirPath, cleanPath)
	}

	relCheck, err := filepath.Rel(dirPath, fullPath)
	if err != nil || relCheck == ".." || strings.HasPrefix(relCheck, ".."+string(os.PathSeparator)) {
		return "", "", errors.New("invalid path")
	}
	return fullPath, dirPath, nil
}

func ensureVideoFileExists(c *gin.Context, fullPath string) error {
	if _, err := os.Stat(fullPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.Status(http.StatusNotFound)
			return err
		}
		logging.Error("stat stream file error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return err
	}
	return nil
}

func incrementPlayCountByPath(ctx context.Context, dirPath, fullPath string) {
	if strings.TrimSpace(dirPath) == "" || strings.TrimSpace(fullPath) == "" {
		return
	}
	relPath, err := filepath.Rel(dirPath, fullPath)
	if err != nil {
		logging.Error("resolve relative path for play count: %v", err)
		return
	}
	relPath = filepath.ToSlash(filepath.Clean(relPath))
	if relPath == "." || strings.HasPrefix(relPath, "..") {
		return
	}
	if err := dbpkg.IncrementVideoPlayCountByPath(ctx, dirPath, relPath); err != nil {
		logging.Error("increment play count by path error: %v", err)
	}
}

func getThumbnail(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	video, err := dbpkg.GetVideo(c.Request.Context(), id)
	if err != nil {
		logging.Error("get screenshot error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if video == nil {
		c.Status(http.StatusNotFound)
		return
	}

	second, ok := manager.PickScreenshotSecond(video.DurationSec)
	if !ok {
		c.Status(http.StatusNotFound)
		return
	}

	if common.AppConfig == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	dataDir := filepath.Dir(common.AppConfig.DatabasePath)
	screenshotPath := manager.ScreenshotPath(dataDir, video.ID, second)
	if screenshotPath == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	if _, err := os.Stat(screenshotPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			common.ScreenshotManager.EnqueueForVideo(video)
			c.Status(http.StatusNotFound)
			return
		}
		logging.Error("stat screenshot error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.File(screenshotPath)
}
