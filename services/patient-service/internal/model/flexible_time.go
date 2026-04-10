package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// FlexibleTime unmarshals either RFC3339 timestamps or date-only strings (YYYY-MM-DD).
// This is useful when the API stores a DATE but clients naturally send only a date.
type FlexibleTime struct {
	time.Time
}

func (t *FlexibleTime) UnmarshalJSON(data []byte) error {
	// For pointer fields, JSON null is handled by the standard library (pointer becomes nil),
	// but keep this here for completeness if used as a non-pointer field.
	if bytes.Equal(data, []byte("null")) {
		*t = FlexibleTime{}
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("FlexibleTime: expected string: %w", err)
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return fmt.Errorf("FlexibleTime: empty string")
	}

	// Try RFC3339 first, then date-only.
	if parsed, err := time.Parse(time.RFC3339, s); err == nil {
		t.Time = parsed
		return nil
	}
	if parsed, err := time.Parse("2006-01-02", s); err == nil {
		t.Time = parsed
		return nil
	}

	return fmt.Errorf("FlexibleTime: invalid format %q (use RFC3339 or YYYY-MM-DD)", s)
}

