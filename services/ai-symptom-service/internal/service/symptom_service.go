package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"healthcare-platform/pkg/logger"
	"healthcare-platform/services/ai-symptom-service/internal/integrations"
	"healthcare-platform/services/ai-symptom-service/internal/model"
)

const medicalDisclaimer = "This information is not a medical diagnosis. Always consult a qualified healthcare professional for evaluation and treatment."

const systemPrompt = `You are a clinical triage assistant for a healthcare platform. You must NOT provide a definitive diagnosis or prescribe medication.
Your task is to suggest which medical specialty is most appropriate for follow-up and to give brief, plain-language preliminary notes that help the patient understand possible next steps.
Respond ONLY with a single JSON object (no markdown) using exactly these keys:
- "suggested_specialty": a short specialty name (e.g. "Cardiology", "General practice").
- "preliminary_notes": 2–5 sentences, cautious wording, encourage professional care when appropriate.
If information is insufficient, say so in preliminary_notes and suggest general practice or emergency care if red-flag symptoms appear.`

var (
	ErrLLMOutputInvalid    = errors.New("model returned invalid JSON structure")
	ErrLLMQuotaOrRateLimit = errors.New("LLM quota or rate limit exceeded")
)

// SymptomService orchestrates symptom analysis via the configured LLM (stateless).
type SymptomService struct {
	llm integrations.LLMClient
	log *logger.Logger
}

func NewSymptomService(llm integrations.LLMClient, log *logger.Logger) *SymptomService {
	return &SymptomService{llm: llm, log: log}
}

// Check analyzes reported symptoms for a patient (non-diagnostic).
func (s *SymptomService) Check(ctx context.Context, req *model.SymptomCheckRequest) (*model.SymptomCheckResponse, error) {
	user := buildUserPrompt(req)
	raw, err := s.llm.CompleteJSON(ctx, systemPrompt, user)
	if err != nil {
		if isLLMQuotaOrRateLimit(err) {
			return nil, ErrLLMQuotaOrRateLimit
		}
		return nil, fmt.Errorf("llm complete: %w", err)
	}

	var parsed struct {
		SuggestedSpecialty string `json:"suggested_specialty"`
		PreliminaryNotes   string `json:"preliminary_notes"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		s.log.Warn("LLM JSON parse failed", "error", err, "snippet", truncate(string(raw), 300))
		return nil, fmt.Errorf("%w: %v", ErrLLMOutputInvalid, err)
	}

	parsed.SuggestedSpecialty = strings.TrimSpace(parsed.SuggestedSpecialty)
	parsed.PreliminaryNotes = strings.TrimSpace(parsed.PreliminaryNotes)
	if parsed.SuggestedSpecialty == "" || parsed.PreliminaryNotes == "" {
		return nil, ErrLLMOutputInvalid
	}

	return &model.SymptomCheckResponse{
		SuggestedSpecialty: parsed.SuggestedSpecialty,
		PreliminaryNotes:   parsed.PreliminaryNotes,
		Disclaimer:         medicalDisclaimer,
	}, nil
}

func buildUserPrompt(req *model.SymptomCheckRequest) string {
	var b strings.Builder
	b.WriteString("Patient-reported symptoms:\n")
	b.WriteString(strings.TrimSpace(req.Symptoms))
	if strings.TrimSpace(req.OptionalContext) != "" {
		b.WriteString("\n\nAdditional context:\n")
		b.WriteString(strings.TrimSpace(req.OptionalContext))
	}
	return b.String()
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func isLLMQuotaOrRateLimit(err error) bool {
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "quota") ||
		strings.Contains(s, "resource_exhausted") ||
		strings.Contains(s, "rate limit") ||
		strings.Contains(s, "too many requests")
}
