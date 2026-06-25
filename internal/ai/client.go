package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type Client struct {
	apiKey      string
	baseURL     string
	model       string
	visionModel string
	http        *http.Client
}

func NewClient(apiKey, baseURL, model, visionModel string) *Client {
	baseURL = normalizeBaseURL(baseURL)
	if strings.TrimSpace(visionModel) == "" {
		visionModel = "qwen-vl-max"
	}
	return &Client{
		apiKey:      apiKey,
		baseURL:     baseURL,
		model:       model,
		visionModel: visionModel,
		http:        &http.Client{Timeout: 180 * time.Second},
	}
}

// normalizeBaseURL 将已废弃的 text-generation 端点映射到 OpenAI 兼容模式（qwen3 等新模型必需）
func normalizeBaseURL(url string) string {
	if strings.Contains(url, "text-generation/generation") {
		return "https://dashscope.aliyuncs.com/compatible-mode/v1/chat/completions"
	}
	return url
}

func (c *Client) useCompatibleMode() bool {
	return strings.Contains(c.baseURL, "compatible-mode") ||
		strings.Contains(c.baseURL, "chat/completions")
}

func (c *Client) Complete(prompt string) (string, error) {
	return c.CompleteWithImages(prompt, nil)
}

// CompleteWithImages 文本 + 设计稿截图（data URL 或 http URL）多模态调用。
func (c *Client) CompleteWithImages(prompt string, imageDataURLs []string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("AI_API_KEY not configured")
	}
	if c.useCompatibleMode() {
		return c.completeCompatibleMultimodal(prompt, imageDataURLs)
	}
	if len(imageDataURLs) > 0 {
		return "", fmt.Errorf("设计稿截图需要配置 OpenAI 兼容模式 AI_BASE_URL（含 chat/completions）")
	}
	return c.completeLegacy(prompt)
}

// completeCompatible OpenAI 兼容模式（推荐，支持 qwen-plus / qwen-max 等）
func (c *Client) completeCompatible(prompt string) (string, error) {
	return c.completeCompatibleMultimodal(prompt, nil)
}

func (c *Client) completeCompatibleMultimodal(prompt string, imageDataURLs []string) (string, error) {
	model := c.model
	var content any = prompt
	if len(imageDataURLs) > 0 {
		model = c.visionModel
		parts := []map[string]any{{"type": "text", "text": prompt}}
		for _, img := range imageDataURLs {
			if strings.TrimSpace(img) == "" {
				continue
			}
			parts = append(parts, map[string]any{
				"type": "image_url",
				"image_url": map[string]string{"url": img},
			})
		}
		content = parts
	}
	body := map[string]any{
		"model": model,
		"messages": []map[string]any{
			{"role": "user", "content": content},
		},
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("AI HTTP %d: %s", resp.StatusCode, string(b))
	}

	// OpenAI 格式
	var openAI struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(b, &openAI); err == nil && len(openAI.Choices) > 0 {
		text := strings.TrimSpace(openAI.Choices[0].Message.Content)
		if text != "" {
			return text, nil
		}
	}
	if openAI.Error != nil && openAI.Error.Message != "" {
		return "", fmt.Errorf("AI error: %s", openAI.Error.Message)
	}

	// DashScope 兼容模式有时仍包一层 output
	var wrapped struct {
		Output struct {
			Text string `json:"text"`
		} `json:"output"`
	}
	if err := json.Unmarshal(b, &wrapped); err == nil && wrapped.Output.Text != "" {
		return strings.TrimSpace(wrapped.Output.Text), nil
	}

	return "", fmt.Errorf("AI response parse failed: %s", truncate(string(b), 300))
}

// completeLegacy 旧版 DashScope text-generation 端点（仅适用于 qwen-plus 等老模型）
func (c *Client) completeLegacy(prompt string) (string, error) {
	body := map[string]any{
		"model": c.model,
		"input": map[string]any{
			"messages": []map[string]string{
				{"role": "user", "content": prompt},
			},
		},
	}
	raw, _ := json.Marshal(body)
	req, err := http.NewRequest(http.MethodPost, c.baseURL, bytes.NewReader(raw))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("AI HTTP %d: %s", resp.StatusCode, string(b))
	}

	var result struct {
		Output struct {
			Text string `json:"text"`
		} `json:"output"`
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return "", fmt.Errorf("parse AI response: %w", err)
	}
	text := strings.TrimSpace(result.Output.Text)
	if text == "" {
		return "", fmt.Errorf("AI returned empty text")
	}
	return text, nil
}

func ParseJSON[T any](raw string, out *T) error {
	cleaned := cleanJSON(raw)
	if err := json.Unmarshal([]byte(cleaned), out); err != nil {
		return fmt.Errorf("invalid AI JSON: %w", err)
	}
	return nil
}

func cleanJSON(s string) string {
	s = strings.TrimSpace(s)
	s = regexp.MustCompile("(?i)^```json\\s*").ReplaceAllString(s, "")
	s = regexp.MustCompile("```\\s*$").ReplaceAllString(s, "")
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		s = s[start : end+1]
	}
	return s
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
