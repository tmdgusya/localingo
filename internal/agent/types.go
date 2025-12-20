package agent

// GenerateRequest represents a request to generate text from an LLM
type GenerateRequest struct {
	Model  string
	Prompt string
	System string
}

// GenerateResponse represents a response from an LLM
type GenerateResponse struct {
	Text string
	Done bool
}
