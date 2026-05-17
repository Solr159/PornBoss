package deepseek

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.deepseek.com"
const defaultModel = "deepseek-chat"

// Client calls DeepSeek-compatible chat completions APIs.
type Client struct {
	HTTPClient *http.Client
	BaseURL    string
	APIKey     string
	Model      string
}

func (c *Client) base() string {
	b := strings.TrimSpace(c.BaseURL)
	if b == "" {
		b = defaultBaseURL
	}
	return strings.TrimRight(b, "/")
}

func (c *Client) model() string {
	m := strings.TrimSpace(c.Model)
	if m == "" {
		return defaultModel
	}
	return m
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatMessage is one message in a multi-turn chat completion.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

// ChatCompletion returns the assistant message content (non-streaming).
func (c *Client) ChatCompletion(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return c.ChatCompletionTemperature(ctx, systemPrompt, userPrompt, 0.2, 2048)
}

// ChatCompletionTemperature posts a two-message completion with custom temperature and max tokens.
func (c *Client) ChatCompletionTemperature(ctx context.Context, systemPrompt, userPrompt string, temperature float64, maxTokens int) (string, error) {
	key := strings.TrimSpace(c.APIKey)
	if key == "" {
		return "", errors.New("deepseek api key is not configured")
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 120 * time.Second}
	}
	if temperature <= 0 {
		temperature = 0.2
	}
	if maxTokens <= 0 {
		maxTokens = 2048
	}
	body := chatRequest{
		Model: c.model(),
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("deepseek http %d: %s", resp.StatusCode, string(respBody))
	}
	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decode deepseek response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", errors.New(parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", errors.New("empty deepseek response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}

// ChatMessages posts a chat completion with the given message list (roles: system, user, assistant).
func (c *Client) ChatMessages(ctx context.Context, messages []ChatMessage) (string, error) {
	return c.ChatMessagesTemperature(ctx, messages, 0)
}

// ChatMessagesTemperature is like ChatMessages but sets sampling temperature; values <= 0 use 0.6.
func (c *Client) ChatMessagesTemperature(ctx context.Context, messages []ChatMessage, temperature float64) (string, error) {
	key := strings.TrimSpace(c.APIKey)
	if key == "" {
		return "", errors.New("deepseek api key is not configured")
	}
	if len(messages) == 0 {
		return "", errors.New("no messages")
	}
	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: 120 * time.Second}
	}
	if temperature <= 0 {
		temperature = 0.6
	}
	if temperature > 1.5 {
		temperature = 1.5
	}
	out := make([]chatMessage, 0, len(messages))
	for _, m := range messages {
		role := strings.TrimSpace(m.Role)
		content := strings.TrimSpace(m.Content)
		if role == "" || content == "" {
			continue
		}
		out = append(out, chatMessage{Role: role, Content: content})
	}
	if len(out) == 0 {
		return "", errors.New("no valid messages")
	}
	body := chatRequest{
		Model:       c.model(),
		Messages:    out,
		Temperature: temperature,
		MaxTokens:   4096,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base()+"/v1/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("deepseek http %d: %s", resp.StatusCode, string(respBody))
	}
	var parsed chatResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", fmt.Errorf("decode deepseek response: %w", err)
	}
	if parsed.Error != nil && parsed.Error.Message != "" {
		return "", errors.New(parsed.Error.Message)
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return "", errors.New("empty deepseek response")
	}
	return strings.TrimSpace(parsed.Choices[0].Message.Content), nil
}
