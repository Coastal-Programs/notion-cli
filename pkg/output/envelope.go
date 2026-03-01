package output

import "time"

// Version is the CLI version included in envelope metadata.
const Version = "6.0.0"

// SuccessEnvelope wraps successful command output.
type SuccessEnvelope struct {
	Success  bool           `json:"success"`
	Data     any            `json:"data"`
	Metadata map[string]any `json:"metadata"`
}

// ErrorDetail contains structured error information.
type ErrorDetail struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Details     any      `json:"details,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// ErrorEnvelope wraps error output.
type ErrorEnvelope struct {
	Success  bool           `json:"success"`
	Error    ErrorDetail    `json:"error"`
	Metadata map[string]any `json:"metadata"`
}

// NewSuccessEnvelope creates an envelope for successful output.
func NewSuccessEnvelope(data any, command string, executionMs int64) *SuccessEnvelope {
	return &SuccessEnvelope{
		Success: true,
		Data:    data,
		Metadata: map[string]any{
			"timestamp":         time.Now().UTC().Format(time.RFC3339),
			"command":           command,
			"execution_time_ms": executionMs,
			"version":           Version,
		},
	}
}

// NewErrorEnvelope creates an envelope for error output.
func NewErrorEnvelope(code, message string, details any, suggestions []string, command string) *ErrorEnvelope {
	return &ErrorEnvelope{
		Success: false,
		Error: ErrorDetail{
			Code:        code,
			Message:     message,
			Details:     details,
			Suggestions: suggestions,
		},
		Metadata: map[string]any{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"command":   command,
			"version":   Version,
		},
	}
}
