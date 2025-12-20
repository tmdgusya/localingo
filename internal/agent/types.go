package agent

import (
	"encoding/json"
	"strings"
)

// GenerateRequest represents a request to generate text from an LLM
type GenerateRequest struct {
	Model  string
	Prompt string
	System string
}

// Analysis represents the linguistic analysis of the user's input
type Analysis struct {
	Categories  []string `json:"categories"`
	Explanation string   `json:"explanation"`
}

// StructuredResponse is the expected JSON format from the LLM
type StructuredResponse struct {
	Rephrased string   `json:"rephrased"`
	Analysis  Analysis `json:"analysis"`
}

// GenerateResponse represents a response from an LLM, now with optional structured data
type GenerateResponse struct {
	Text     string   // The rephrased text (backward compatibility)
	Done     bool     // Stream status
	Analysis *Analysis // Structured analysis data (optional)
}

// ParseStructuredResponse attempts to parse a string into a StructuredResponse
// Returns a partial response if parsing fails but text exists
func ParseStructuredResponse(text string) *GenerateResponse {
	// Clean up potential markdown code blocks (e.g. ```json ... ```)
	cleanText := text
	cleanText = strings.TrimPrefix(cleanText, "```json")
	cleanText = strings.TrimPrefix(cleanText, "```")
	cleanText = strings.TrimSuffix(cleanText, "```")
	cleanText = strings.TrimSpace(cleanText)

	var structResp StructuredResponse
	if err := json.Unmarshal([]byte(cleanText), &structResp); err != nil {
		// Fallback: If not valid JSON, treat entire text as rephrased content
		return &GenerateResponse{
			Text: text,
			Done: true,
		}
	}

	return &GenerateResponse{
		Text:     structResp.Rephrased,
		Done:     true,
		Analysis: &structResp.Analysis,
	}
}

// CorrectionLog represents a past correction for analysis
type CorrectionLog struct {
	Original  string
	Corrected string
	Reason    string
}