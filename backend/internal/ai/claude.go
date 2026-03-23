package ai

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/skriptvalley/careerdock/internal/ai/prompts"
)

const (
	claudeAPIURL     = "https://api.anthropic.com/v1/messages"
	claudeAPIVersion = "2023-06-01"
	defaultModel     = "claude-sonnet-4-20250514"
	defaultMaxTokens = 4096
)

// ClaudeProvider implements LLMProvider using the Anthropic Claude API.
type ClaudeProvider struct {
	apiKey    string
	model     string
	maxTokens int
	client    *http.Client
}

// NewClaudeProvider creates a new Claude API provider.
func NewClaudeProvider(apiKey, model string, maxTokens int) *ClaudeProvider {
	if model == "" {
		model = defaultModel
	}
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}
	return &ClaudeProvider{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		client:    &http.Client{Timeout: 90 * time.Second},
	}
}

// Name returns the provider name.
func (c *ClaudeProvider) Name() string { return "claude" }

// ParseResume extracts structured data from resume text using Claude.
func (c *ClaudeProvider) ParseResume(ctx context.Context, req *ParseResumeRequest) (*ParsedResume, error) {
	systemPrompt := prompts.BuildSystemPrompt(prompts.ResumeParseSystem())
	userMessage := prompts.ResumeParseUser(req.ResumeText)

	raw, tokens, err := c.callText(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, fmt.Errorf("claude parse resume: %w", err)
	}

	var result ParsedResume
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("claude parse resume JSON: %w (raw: %s)", err, truncate(raw, 200))
	}

	result.TokensUsed = tokens
	return &result, nil
}

// ScoreATSGeneral evaluates a resume's general ATS compatibility.
// Uses the raw PDF bytes for document-level analysis when available.
func (c *ClaudeProvider) ScoreATSGeneral(ctx context.Context, req *ATSGeneralRequest) (*ATSResult, error) {
	systemPrompt := prompts.BuildSystemPrompt(prompts.ATSGeneralSystem())

	var raw []byte
	var tokens TokenUsage
	var err error

	if len(req.PDFBytes) > 0 {
		userMessage := prompts.ATSGeneralUserPDF()
		raw, tokens, err = c.callWithPDF(ctx, systemPrompt, userMessage, req.PDFBytes)
	} else {
		userMessage := prompts.ATSGeneralUserText(req.ResumeText)
		raw, tokens, err = c.callText(ctx, systemPrompt, userMessage)
	}
	if err != nil {
		return nil, fmt.Errorf("claude ATS general: %w", err)
	}

	var result ATSResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("claude ATS general JSON: %w (raw: %s)", err, truncate(raw, 200))
	}

	result.TokensUsed = tokens
	result.GeneratedAt = time.Now()
	return &result, nil
}

// --- Internal HTTP methods ---

// claudeRequest is the request body for the Claude Messages API.
type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeMessage struct {
	Role    string         `json:"role"`
	Content []contentBlock `json:"content"`
}

type contentBlock struct {
	Type   string       `json:"type"`
	Text   string       `json:"text,omitempty"`
	Source *sourceBlock `json:"source,omitempty"`
}

type sourceBlock struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

// claudeResponse is the response from the Claude Messages API.
type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// callText sends a text-only request to Claude.
func (c *ClaudeProvider) callText(ctx context.Context, system, userText string) ([]byte, TokenUsage, error) {
	body := claudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    system,
		Messages: []claudeMessage{
			{
				Role: "user",
				Content: []contentBlock{
					{Type: "text", Text: userText},
				},
			},
		},
	}

	return c.doRequest(ctx, body)
}

// callWithPDF sends a request with an embedded PDF document to Claude.
func (c *ClaudeProvider) callWithPDF(ctx context.Context, system, userText string, pdfBytes []byte) ([]byte, TokenUsage, error) {
	encoded := base64.StdEncoding.EncodeToString(pdfBytes)

	body := claudeRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		System:    system,
		Messages: []claudeMessage{
			{
				Role: "user",
				Content: []contentBlock{
					{
						Type: "document",
						Source: &sourceBlock{
							Type:      "base64",
							MediaType: "application/pdf",
							Data:      encoded,
						},
					},
					{Type: "text", Text: userText},
				},
			},
		},
	}

	return c.doRequest(ctx, body)
}

// doRequest executes an HTTP request to the Claude API and extracts the text response.
func (c *ClaudeProvider) doRequest(ctx context.Context, body claudeRequest) ([]byte, TokenUsage, error) {
	payload, err := json.Marshal(body)
	if err != nil {
		return nil, TokenUsage{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, claudeAPIURL, bytes.NewReader(payload))
	if err != nil {
		return nil, TokenUsage{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", claudeAPIVersion)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, TokenUsage{}, fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, TokenUsage{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("claude API error",
			"status", resp.StatusCode,
			"body", truncate(respBody, 500),
		)
		return nil, TokenUsage{}, fmt.Errorf("claude API error (status %d): %s", resp.StatusCode, truncate(respBody, 200))
	}

	var claudeResp claudeResponse
	if err := json.Unmarshal(respBody, &claudeResp); err != nil {
		return nil, TokenUsage{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if claudeResp.Error != nil {
		return nil, TokenUsage{}, fmt.Errorf("claude error: %s — %s", claudeResp.Error.Type, claudeResp.Error.Message)
	}

	// Extract text from first text content block
	for _, block := range claudeResp.Content {
		if block.Type == "text" {
			tokens := TokenUsage{
				InputTokens:  claudeResp.Usage.InputTokens,
				OutputTokens: claudeResp.Usage.OutputTokens,
			}
			// Strip markdown code fences if present
			text := stripCodeFences(block.Text)
			return []byte(text), tokens, nil
		}
	}

	return nil, TokenUsage{}, fmt.Errorf("no text content in Claude response")
}

// truncate shortens a byte slice for log output.
func truncate(b []byte, maxLen int) string {
	s := string(b)
	if len(s) > maxLen {
		return s[:maxLen] + "..."
	}
	return s
}

// stripCodeFences removes markdown ```json ... ``` wrapping from LLM output.
func stripCodeFences(s string) string {
	// Trim surrounding whitespace first
	trimmed := []byte(s)
	trimmed = bytes.TrimSpace(trimmed)

	// Check for ```json or ``` prefix
	if bytes.HasPrefix(trimmed, []byte("```json")) {
		trimmed = trimmed[7:]
	} else if bytes.HasPrefix(trimmed, []byte("```")) {
		trimmed = trimmed[3:]
	}

	// Check for ``` suffix
	if bytes.HasSuffix(trimmed, []byte("```")) {
		trimmed = trimmed[:len(trimmed)-3]
	}

	return string(bytes.TrimSpace(trimmed))
}
