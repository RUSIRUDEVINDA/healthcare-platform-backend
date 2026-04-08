package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const openAIChatURL = "https://api.openai.com/v1/chat/completions"

// OpenAIClient calls OpenAI Chat Completions with JSON response format.
type OpenAIClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewOpenAIClient(apiKey, model string) *OpenAIClient {
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIClient{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 90 * time.Second},
	}
}

type openAIRequest struct {
	Model          string              `json:"model"`
	Messages       []openAIMessage     `json:"messages"`
	ResponseFormat *openAIRespFormat   `json:"response_format,omitempty"`
	Temperature    float64             `json:"temperature"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIRespFormat struct {
	Type string `json:"type"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *OpenAIClient) CompleteJSON(ctx context.Context, systemPrompt, userPrompt string) ([]byte, error) {
	body := openAIRequest{
		Model: c.model,
		Messages: []openAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		ResponseFormat: &openAIRespFormat{Type: "json_object"},
		Temperature:    0.3,
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, openAIChatURL, bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openai request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed openAIResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("openai decode: %w", bodyPreviewErr(respBody, err))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("openai api: %s", parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("openai http %d: %s", resp.StatusCode, truncate(string(respBody), 500))
	}
	if len(parsed.Choices) == 0 || parsed.Choices[0].Message.Content == "" {
		return nil, fmt.Errorf("openai: empty completion")
	}
	return []byte(parsed.Choices[0].Message.Content), nil
}

func bodyPreviewErr(b []byte, err error) error {
	return fmt.Errorf("%w (body=%q)", err, truncate(string(b), 200))
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
