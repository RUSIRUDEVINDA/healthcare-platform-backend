package integrations

import "context"

// LLMClient requests a JSON object from a provider (OpenAI or Gemini).
// Each implementation returns the raw JSON text from the model response body.
type LLMClient interface {
	CompleteJSON(ctx context.Context, systemPrompt, userPrompt string) ([]byte, error)
}
