package integrations

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// GeminiClient calls Google Generative Language API (generateContent).
type GeminiClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewGeminiClient(apiKey, model string) *GeminiClient {
	if model == "" {
		model = "gemini-2.5-flash"
	}
	return &GeminiClient{
		apiKey: apiKey,
		model:  model,
		http:   &http.Client{Timeout: 90 * time.Second},
	}
}

type geminiRequest struct {
	SystemInstruction *geminiContent `json:"systemInstruction,omitempty"`
	Contents          []geminiTurn   `json:"contents"`
	GenerationConfig  geminiGenCfg   `json:"generationConfig"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiTurn struct {
	Role  string         `json:"role"`
	Parts []geminiPart   `json:"parts"`
}

type geminiGenCfg struct {
	ResponseMIMEType string  `json:"responseMimeType"`
	Temperature      float64 `json:"temperature"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []geminiPart `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (c *GeminiClient) CompleteJSON(ctx context.Context, systemPrompt, userPrompt string) ([]byte, error) {
	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent",
		url.PathEscape(c.model),
	)
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("key", c.apiKey)
	u.RawQuery = q.Encode()

	reqBody := geminiRequest{
		SystemInstruction: &geminiContent{Parts: []geminiPart{{Text: systemPrompt}}},
		Contents: []geminiTurn{
			{Role: "user", Parts: []geminiPart{{Text: userPrompt}}},
		},
		GenerationConfig: geminiGenCfg{
			ResponseMIMEType: "application/json",
			Temperature:      0.3,
		},
	}
	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var parsed geminiResponse
	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, fmt.Errorf("gemini decode: %w", bodyPreviewErr(respBody, err))
	}
	if parsed.Error != nil {
		return nil, fmt.Errorf("gemini api: %s", parsed.Error.Message)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("gemini http %d: %s", resp.StatusCode, truncate(string(respBody), 500))
	}
	if len(parsed.Candidates) == 0 || len(parsed.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini: empty completion")
	}
	text := parsed.Candidates[0].Content.Parts[0].Text
	if text == "" {
		return nil, fmt.Errorf("gemini: empty text")
	}
	return []byte(text), nil
}
