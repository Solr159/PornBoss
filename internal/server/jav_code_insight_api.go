package server

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common/logging"
	"pornboss/internal/deepseek"
)

const javCodeInsightJSONSchema = `{
  "code": "IPX-811",
  "title_cn": "中文标题（如果没有就用日文）",
  "title_jp": "原日文标题",
  "actress": ["楓カレン", "枫可怜"],
  "studio": "IDEA Pocket",
  "release_date": "2022-02-XX",
  "duration": "114",
  "genres": ["媚药", "相部屋", "NTR", "出差", "哭脸", "Ahegao"],
  "plot_summary": "简短剧情总结（100-150字，突出核心冲突和高潮部分）",
  "rating": 8.7,
  "highlights": ["女优巅峰表现", "媚药觉醒过程极色", "后半段持久战"],
  "weaknesses": ["男优一般", "部分镜头重复"],
  "similar_codes": ["IPX-XXX", "IPX-YYY"],
  "cover_desc": "封面描述（可选）"
}`

// JavCodeInsight is the structured metadata returned for a catalog code.
type JavCodeInsight struct {
	Code         string   `json:"code"`
	TitleCN      string   `json:"title_cn"`
	TitleJP      string   `json:"title_jp"`
	Actress      []string `json:"actress"`
	Studio       string   `json:"studio"`
	ReleaseDate  string   `json:"release_date"`
	Duration     string   `json:"duration"`
	Genres       []string `json:"genres"`
	PlotSummary  string   `json:"plot_summary"`
	Rating       float64  `json:"rating"`
	Highlights   []string `json:"highlights"`
	Weaknesses   []string `json:"weaknesses"`
	SimilarCodes []string `json:"similar_codes"`
	CoverDesc    string   `json:"cover_desc"`
}

func javCodeInsightSystemPrompt() string {
	return `你是一个专业的 JAV 成人视频元数据专家，熟悉所有主流制作商（IDEA Pocket、S1、Prestige、Madonna 等）和女优信息。
你只能根据番号与行业公开常识做推断；若不确定请合理推测并在字段中体现不确定性，但不要声称读取了用户本地数据库或文件。
输出必须是合法 JSON 对象，不要添加任何其他文字、解释或 Markdown 代码块。`
}

func postJavCodeInsight(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	code := strings.TrimSpace(req.Code)
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code is required"})
		return
	}
	if len(code) > 64 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code too long"})
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

	user := "给定番号：" + code + `

请严格以以下 JSON 格式返回，不要添加任何其他文字、解释或 Markdown：

` + javCodeInsightJSONSchema

	client := &deepseek.Client{APIKey: apiKey, BaseURL: baseURL}
	raw, err := client.ChatCompletionTemperature(c.Request.Context(), javCodeInsightSystemPrompt(), user, 0.25, 4096)
	if err != nil {
		logging.Error("jav code insight: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	jsonStr, err := extractJSONObjectFromLLM(raw)
	if err != nil {
		logging.Error("jav code insight json: %v raw=%q", err, raw)
		c.JSON(http.StatusBadGateway, gin.H{"error": "model returned invalid format"})
		return
	}
	var insight JavCodeInsight
	if err := json.Unmarshal([]byte(jsonStr), &insight); err != nil {
		logging.Error("jav code insight parse: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": "model json parse failed"})
		return
	}
	if strings.TrimSpace(insight.Code) == "" {
		insight.Code = code
	}
	c.JSON(http.StatusOK, gin.H{"insight": insight})
}
