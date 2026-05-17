package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"

	"pornboss/internal/common/logging"
	dbpkg "pornboss/internal/db"
	"pornboss/internal/deepseek"
)

func buildJavLibrarySnapshotText(ctx context.Context, directoryIDs []int64, collectionID int64) (string, error) {
	_, total, err := dbpkg.SearchJav(ctx, nil, nil, "", "recent", 1, 0, nil, directoryIDs, collectionID, 0)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "visible_work_total: %d\n", total)
	if collectionID > 0 {
		col, err := dbpkg.GetCollection(ctx, collectionID)
		if err == nil && col != nil {
			fmt.Fprintf(&b, "scoped_collection: %s (id=%d)\n", strings.TrimSpace(col.Name), col.ID)
		} else {
			fmt.Fprintf(&b, "scoped_collection_id: %d\n", collectionID)
		}
	}
	tags, err := dbpkg.ListJavTags(ctx, directoryIDs)
	if err != nil {
		return "", err
	}
	type tagRow struct {
		name  string
		count int64
	}
	var tagRows []tagRow
	for _, t := range tags {
		if t.Count <= 0 || strings.TrimSpace(t.Name) == "" {
			continue
		}
		tagRows = append(tagRows, tagRow{name: strings.TrimSpace(t.Name), count: t.Count})
	}
	sort.Slice(tagRows, func(i, j int) bool {
		if tagRows[i].count != tagRows[j].count {
			return tagRows[i].count > tagRows[j].count
		}
		return tagRows[i].name < tagRows[j].name
	})
	if len(tagRows) > 20 {
		tagRows = tagRows[:20]
	}
	if len(tagRows) > 0 {
		b.WriteString("top_tags: ")
		parts := make([]string, 0, len(tagRows))
		for _, r := range tagRows {
			parts = append(parts, fmt.Sprintf("%s(%d)", r.name, r.count))
		}
		b.WriteString(strings.Join(parts, ", "))
		b.WriteString("\n")
	}
	idols, _, err := dbpkg.ListJavIdols(ctx, "", "work", 15, 0, directoryIDs)
	if err != nil {
		return "", err
	}
	if len(idols) > 0 {
		b.WriteString("top_idols: ")
		parts := make([]string, 0, len(idols))
		for _, ido := range idols {
			n := strings.TrimSpace(ido.Name)
			if n == "" {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s(%d)", n, ido.WorkCount))
		}
		b.WriteString(strings.Join(parts, ", "))
		b.WriteString("\n")
	}
	studios, _, err := dbpkg.ListJavStudios(ctx, "", 10, 0, directoryIDs)
	if err != nil {
		return "", err
	}
	if len(studios) > 0 {
		b.WriteString("top_studios: ")
		parts := make([]string, 0, len(studios))
		for _, s := range studios {
			n := strings.TrimSpace(s.Name)
			if n == "" {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s(%d)", n, s.WorkCount))
		}
		b.WriteString(strings.Join(parts, ", "))
		b.WriteString("\n")
	}
	cols, err := dbpkg.ListCollections(ctx)
	if err != nil {
		return "", err
	}
	if len(cols) > 0 {
		b.WriteString("collections: ")
		parts := make([]string, 0, len(cols))
		for i, c := range cols {
			if i >= 12 {
				break
			}
			n := strings.TrimSpace(c.Name)
			if n == "" {
				continue
			}
			parts = append(parts, fmt.Sprintf("%s(%d)", n, c.Count))
		}
		b.WriteString(strings.Join(parts, ", "))
		b.WriteString("\n")
	}
	return strings.TrimSpace(b.String()), nil
}

func postJavLibraryChat(c *gin.Context) {
	var req struct {
		Message string `json:"message"`
		History []struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"history"`
		DirectoryIDs []int64 `json:"directory_ids"`
		CollectionID int64   `json:"collection_id"`
		// Mode: "" or "library" = aggregate snapshot from DB; "code_reasoning" = model-only speculation from user text (no snapshot).
		Mode      string `json:"mode"`
		FocusCode string `json:"focus_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
		return
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
		return
	}
	if len(msg) > 4000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "message too long"})
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
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode == "" {
		mode = "library"
	}
	var snapshot string
	if mode != "code_reasoning" {
		var err error
		snapshot, err = buildJavLibrarySnapshotText(c.Request.Context(), req.DirectoryIDs, req.CollectionID)
		if err != nil {
			logging.Error("jav library snapshot: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	var system string
	temp := 0.72
	switch mode {
	case "code_reasoning":
		temp = 0.85
		system = `你是熟悉成人影片发行与命名惯例的助手。用户会给出番号（可能写在消息里或单独字段）并提出问题。
重要规则：
1) 你只能基于公开可查的**一般性行业常识**做合理推测（例如从番号前缀推断常见片商/系列风格、题材倾向），语气自然、像朋友在讨论，不要写成机械条款列举。
2) 明确说明这是推测，可能不准；不要声称读取了用户电脑或本地数据库；不要编造该番号在你这边“确实存在”的细节。
3) 不要提供违法内容、不要指导获取盗版资源；避免露骨描写，侧重类型/片商线/系列定位等元信息层面的讨论。
4) 用户用中文则主要用中文回复；用户用其他语言则尽量用同语言回复。`
	default:
		system = `你是用户本地 JAV 元数据库的助手。用户会基于下面提供的「库统计快照」提问：可能包括评价口味、推荐补片方向、标签/演员/片商结构观察、合集规划等。
规则：
1) 只依据快照中的统计与名称作答，不要编造具体文件路径或未在快照出现的条目。
2) 若信息不足以回答，明确说明缺少什么信息，并给出用户可在应用里如何补充筛选或整理的建议。
3) 保持尊重与中立，避免露骨描写；侧重类型结构、片商与演员分布、收藏规划等。回答尽量口语化、有观点，少用套话。
4) 用户用中文则主要用中文回复；用户用其他语言则尽量用同语言回复。`
	}

	var history []deepseek.ChatMessage
	for _, h := range req.History {
		role := strings.ToLower(strings.TrimSpace(h.Role))
		content := strings.TrimSpace(h.Content)
		if content == "" || len(content) > 6000 {
			continue
		}
		if role != "user" && role != "assistant" {
			continue
		}
		history = append(history, deepseek.ChatMessage{Role: role, Content: content})
	}
	if len(history) > 12 {
		history = history[len(history)-12:]
	}
	userBlock := msg
	if fc := strings.TrimSpace(req.FocusCode); fc != "" {
		userBlock = "番号（请围绕它作答，但仍遵守不要假装读取用户本地库）：" + fc + "\n\n" + userBlock
	}
	if mode == "code_reasoning" {
		userBlock = userBlock + "\n\n（说明：本条对话未附带用户本地库的统计快照；请勿假装看到用户文件或私有元数据。）"
	} else {
		userBlock = userBlock + "\n\n---\n【库统计快照】\n" + snapshot
	}
	msgs := []deepseek.ChatMessage{{Role: "system", Content: system}}
	msgs = append(msgs, history...)
	msgs = append(msgs, deepseek.ChatMessage{Role: "user", Content: userBlock})

	client := &deepseek.Client{APIKey: apiKey, BaseURL: baseURL}
	reply, err := client.ChatMessagesTemperature(c.Request.Context(), msgs, temp)
	if err != nil {
		logging.Error("jav library chat: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"reply": reply})
}
