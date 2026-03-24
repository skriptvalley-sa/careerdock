package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/skriptvalley/careerdock/internal/ai/prompts"
)

const (
	openAIAPIURL       = "https://api.openai.com/v1/chat/completions"
	defaultOpenAIModel = "gpt-4o-mini"
)

// OpenAIProvider implements LLMProvider using the OpenAI Chat Completions API.
// Used as a fallback when Claude is unavailable.
// OpenAI does not support native PDF processing, so all inputs are text-based.
type OpenAIProvider struct {
	apiKey    string
	model     string
	maxTokens int
	client    *http.Client
}

// NewOpenAIProvider creates a new OpenAI API provider.
func NewOpenAIProvider(apiKey, model string, maxTokens int) *OpenAIProvider {
	if model == "" {
		model = defaultOpenAIModel
	}
	if maxTokens <= 0 {
		maxTokens = defaultMaxTokens
	}
	return &OpenAIProvider{
		apiKey:    apiKey,
		model:     model,
		maxTokens: maxTokens,
		client:    &http.Client{Timeout: 90 * time.Second},
	}
}

// Name returns the provider name.
func (o *OpenAIProvider) Name() string { return "openai" }

// ParseResume extracts structured data from resume text using OpenAI.
func (o *OpenAIProvider) ParseResume(ctx context.Context, req *ParseResumeRequest) (*ParsedResume, error) {
	systemPrompt := prompts.BuildSystemPrompt(prompts.ResumeParseSystem())
	userMessage := prompts.ResumeParseUser(req.ResumeText)

	raw, tokens, err := o.call(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, fmt.Errorf("openai parse resume: %w", err)
	}

	var result ParsedResume
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("openai parse resume JSON: %w (raw: %s)", err, truncate(raw, 200))
	}

	result.TokensUsed = tokens
	return &result, nil
}

// ScoreATSGeneral evaluates a resume's general ATS compatibility.
// OpenAI can only use extracted text, not the raw PDF.
func (o *OpenAIProvider) ScoreATSGeneral(ctx context.Context, req *ATSGeneralRequest) (*ATSResult, error) {
	systemPrompt := prompts.BuildSystemPrompt(prompts.ATSGeneralSystem())
	userMessage := prompts.ATSGeneralUserText(req.ResumeText)

	raw, tokens, err := o.call(ctx, systemPrompt, userMessage)
	if err != nil {
		return nil, fmt.Errorf("openai ATS general: %w", err)
	}

	var result ATSResult
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, fmt.Errorf("openai ATS general JSON: %w (raw: %s)", err, truncate(raw, 200))
	}

	result.TokensUsed = tokens
	result.GeneratedAt = time.Now()
	return &result, nil
}

// ScoreATSCompany evaluates resume fit against a specific company profile.
// OpenAI can only use extracted text, not the raw PDF.
func (o *OpenAIProvider) ScoreATSCompany(ctx context.Context, req *ATSCompanyRequest) (*ATSResult, error) {
	systemPrompt := prompts.BuildSystemPrompt(prompts.ATSCompanySystem())

	companyProfile := ""
	if req.Company != nil {
		companyProfile = prompts.CompanyProfileText(
			req.Company.Name,
			req.Company.Size,
			req.Company.TechStack,
			req.Company.Domains,
			req.Company.CompensationTier,
		)
	}

	result, err := ValidateATSResultRetry(ctx, 2, ATSCompanyCategories, func() (*ATSResult, error) {
		userMessage := prompts.ATSCompanyUserText(req.ResumeText, companyProfile)
		raw, tokens, callErr := o.call(ctx, systemPrompt, userMessage)
		if callErr != nil {
			return nil, fmt.Errorf("openai ATS company: %w", callErr)
		}

		var r ATSResult
		if err := json.Unmarshal(raw, &r); err != nil {
			return nil, fmt.Errorf("openai ATS company JSON: %w (raw: %s)", err, truncate(raw, 200))
		}
		r.TokensUsed = tokens
		r.GeneratedAt = time.Now()
		return &r, nil
	})
	return result, err
}

// ScoreATSJob evaluates resume fit against a specific job description.
// OpenAI can only use extracted text, not the raw PDF.
func (o *OpenAIProvider) ScoreATSJob(ctx context.Context, req *ATSJobRequest) (*ATSResult, error) {
	systemPrompt := prompts.BuildSystemPrompt(prompts.ATSJobSystem())

	result, err := ValidateATSResultRetry(ctx, 2, ATSJobCategories, func() (*ATSResult, error) {
		userMessage := prompts.ATSJobUserText(req.ResumeText, req.JobDescription)
		raw, tokens, callErr := o.call(ctx, systemPrompt, userMessage)
		if callErr != nil {
			return nil, fmt.Errorf("openai ATS job: %w", callErr)
		}

		var r ATSResult
		if err := json.Unmarshal(raw, &r); err != nil {
			return nil, fmt.Errorf("openai ATS job JSON: %w (raw: %s)", err, truncate(raw, 200))
		}
		r.TokensUsed = tokens
		r.GeneratedAt = time.Now()
		return &r, nil
	})
	return result, err
}

// --- Internal HTTP methods ---

type openAIRequest struct {
	Model     string          `json:"model"`
	Messages  []openAIMessage `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error"`
}

// call sends a chat completion request to OpenAI.
func (o *OpenAIProvider) call(ctx context.Context, system, userText string) ([]byte, TokenUsage, error) {
	body := openAIRequest{
		Model:     o.model,
		MaxTokens: o.maxTokens,
		Messages: []openAIMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: userText},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, TokenUsage{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIAPIURL, bytes.NewReader(payload))
	if err != nil {
		return nil, TokenUsage{}, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, TokenUsage{}, fmt.Errorf("send request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, TokenUsage{}, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("openai API error",
			"status", resp.StatusCode,
			"body", truncate(respBody, 500),
		)
		return nil, TokenUsage{}, fmt.Errorf("openai API error (status %d): %s", resp.StatusCode, truncate(respBody, 200))
	}

	var oaiResp openAIResponse
	if err := json.Unmarshal(respBody, &oaiResp); err != nil {
		return nil, TokenUsage{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if oaiResp.Error != nil {
		return nil, TokenUsage{}, fmt.Errorf("openai error: %s — %s", oaiResp.Error.Type, oaiResp.Error.Message)
	}

	if len(oaiResp.Choices) == 0 {
		return nil, TokenUsage{}, fmt.Errorf("no choices in OpenAI response")
	}

	tokens := TokenUsage{
		InputTokens:  oaiResp.Usage.PromptTokens,
		OutputTokens: oaiResp.Usage.CompletionTokens,
	}

	text := stripCodeFences(oaiResp.Choices[0].Message.Content)
	return []byte(text), tokens, nil
}
