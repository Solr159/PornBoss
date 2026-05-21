package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/deepseek"
)

type nlJavPlan struct {
	Search         string   `json:"search"`
	TagNames       []string `json:"tag_names"`
	IdolNames      []string `json:"idol_names"`
	StudioName     string   `json:"studio_name"`
	Interpretation string   `json:"interpretation"`
}

func postJavNLQuery(c *gin.Context) {
	var req struct {
		Query        string  `json:"query"`
		DirectoryIDs []int64 `json:"directory_ids"`
		CollectionID int64   `json:"collection_id"`
		Limit        int     `json:"limit"`
		Offset       int     `json:"offset"`
		Sort         string  `json:"sort"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	q := strings.TrimSpace(req.Query)
	if q == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}
	if len(q) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query too long"})
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
	system := `你是本地成人影片库检索助手。用户会用自然语言描述想看的作品。
你必须只输出一个 JSON 对象（不要 markdown，不要解释），字段如下：
{"search":"用于番号或标题关键词搜索的字符串，可为空",
"tag_names":["从用户描述中推断的类型/剧情标签名称，中文或日文，数组可为空"],
"idol_names":["演员姓名数组，可为空"],
"studio_name":"片商名称字符串，没有则空字符串",
"interpretation":"用一两句中文说明你如何把这句话映射成上述检索条件"}
交集规则：若 tag_names 或 idol_names 有多个，表示用户想要同时满足（与现有筛选一致）。
不确定的字段留空数组或空字符串。不要输出除 JSON 以外的任何字符。`
	user := "用户描述：" + q
	client := &deepseek.Client{APIKey: apiKey, BaseURL: baseURL}
	raw, err := client.ChatCompletion(c.Request.Context(), system, user)
	if err != nil {
		logging.Error("deepseek nl: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	jsonStr, err := extractJSONObjectFromLLM(raw)
	if err != nil {
		logging.Error("nl json extract: %v raw=%q", err, raw)
		c.JSON(http.StatusBadGateway, gin.H{"error": "model returned invalid format"})
		return
	}
	var plan nlJavPlan
	if err := json.Unmarshal([]byte(jsonStr), &plan); err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "model json parse failed"})
		return
	}
	tagIDs, tagMiss := dbpkg.ResolveJavTagNamesToIDs(c.Request.Context(), plan.TagNames)
	idolIDs, idolMiss := dbpkg.ResolveJavIdolNamesToIDs(c.Request.Context(), plan.IdolNames)
	studioID := int64(0)
	if strings.TrimSpace(plan.StudioName) != "" {
		sid, err := dbpkg.ResolveStudioNameToID(c.Request.Context(), plan.StudioName)
		if err != nil {
			logging.Error("resolve studio: %v", err)
		} else {
			studioID = sid
		}
	}
	warnings := append([]string{}, tagMiss...)
	warnings = append(warnings, idolMiss...)

	limit := req.Limit
	if limit <= 0 {
		limit = 24
	}
	if limit > 100 {
		limit = 100
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	sort := strings.TrimSpace(req.Sort)
	if sort == "" {
		sort = "recent"
	}
	dirIDs := req.DirectoryIDs
	items, total, err := dbpkg.SearchJav(c.Request.Context(), idolIDs, tagIDs, strings.TrimSpace(plan.Search), sort, limit, offset, nil, dirIDs, req.CollectionID, studioID, 0)
	if err != nil {
		logging.Error("nl SearchJav: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"interpretation": strings.TrimSpace(plan.Interpretation),
		"warnings":       warnings,
		"resolved": gin.H{
			"search":        strings.TrimSpace(plan.Search),
			"tag_ids":       tagIDs,
			"idol_ids":      idolIDs,
			"studio_id":     studioID,
			"collection_id": req.CollectionID,
		},
		"items": items,
		"total": total,
	})
}
