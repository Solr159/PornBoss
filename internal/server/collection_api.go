package server

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/deepseek"
)

func listCollections(c *gin.Context) {
	items, err := dbpkg.ListCollections(c.Request.Context())
	if err != nil {
		logging.Error("list collections: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if items == nil {
		items = []dbpkg.CollectionSummary{}
	}
	c.JSON(http.StatusOK, items)
}

func createCollection(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	col, err := dbpkg.CreateCollection(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		logging.Error("create collection: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, col)
}

func getCollection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	col, err := dbpkg.GetCollection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, col)
}

func patchCollection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	col, err := dbpkg.UpdateCollection(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		logging.Error("update collection: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, col)
}

func deleteCollection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := dbpkg.DeleteCollection(c.Request.Context(), id); err != nil {
		logging.Error("delete collection: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type collectionJavBatchRequest struct {
	CollectionID int64   `json:"collection_id"`
	JavIDs       []int64 `json:"jav_ids"`
}

func addJavsToCollection(c *gin.Context) {
	var req collectionJavBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.CollectionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection_id must be positive"})
		return
	}
	if err := dbpkg.AddJavsToCollection(c.Request.Context(), req.CollectionID, req.JavIDs); err != nil {
		logging.Error("add javs to collection: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func removeJavsFromCollection(c *gin.Context) {
	var req collectionJavBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	if req.CollectionID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "collection_id must be positive"})
		return
	}
	if err := dbpkg.RemoveJavsFromCollection(c.Request.Context(), req.CollectionID, req.JavIDs); err != nil {
		logging.Error("remove javs from collection: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func analyzeCollection(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	apiKey, baseURL, err := deepseekCredentials(c.Request.Context())
	if err != nil {
		logging.Error("deepseek creds: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if apiKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "deepseek api key is not configured"})
		return
	}
	items, err := dbpkg.ListJavsInCollectionForAnalysis(c.Request.Context(), id, 200)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(items) == 0 {
		c.JSON(http.StatusOK, gin.H{"content": "该合集暂无可分析的条目。"})
		return
	}
	var b strings.Builder
	for _, j := range items {
		tagNames := make([]string, 0, len(j.Tags))
		for _, t := range j.Tags {
			if strings.TrimSpace(t.Name) != "" {
				tagNames = append(tagNames, t.Name)
			}
		}
		idolNames := make([]string, 0, len(j.Idols))
		for _, idol := range j.Idols {
			if strings.TrimSpace(idol.Name) != "" {
				idolNames = append(idolNames, idol.Name)
			}
		}
		fmt.Fprintf(&b, "- 番号:%s 标题:%s 演员:%s 类型:%s\n", j.Code, strings.TrimSpace(j.Title+" "+j.TitleEn),
			strings.Join(idolNames, ","), strings.Join(tagNames, ","))
	}
	system := `你是成人影片元数据分析助手。用户会提供一份合集中的作品清单（每行含番号、标题、演员、类型标签）。
请用中文简洁总结：整体题材/类型倾向、出现较多的演员或系列特征、适合怎样口味的观众（一两句即可）。
不要编造清单中不存在的番号；不要输出违法操作建议；不要过度露骨描写。`
	user := "以下为合集中的作品条目：\n" + b.String()
	client := &deepseek.Client{APIKey: apiKey, BaseURL: baseURL}
	out, err := client.ChatCompletion(c.Request.Context(), system, user)
	if err != nil {
		logging.Error("deepseek analyze: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"content": out})
}
