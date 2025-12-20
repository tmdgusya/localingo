package agent

import (
	"context"
	"fmt"

	"github.com/tmdgusya/localingo/internal/prompt"
)

type PhraseSenseiAgent struct {
	client        LLMClient
	promptManager *prompt.Manager
}

func NewPhraseSenseiAgent(client LLMClient, pm *prompt.Manager) *PhraseSenseiAgent {
	return &PhraseSenseiAgent{
		client:        client,
		promptManager: pm,
	}
}

// RephraseAndAnalyze generates a rephrased version and analyzes errors.
// It uses the configured prompt template to enforce JSON output.
func (a *PhraseSenseiAgent) RephraseAndAnalyze(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// 1. Prepare Prompt
	templateName := "default_analyze"
	
	promptText, err := a.promptManager.Execute(templateName, map[string]string{
		"Input": req.Prompt,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to execute prompt template: %w", err)
	}

	// 2. Update Request with formatted prompt
	llmReq := &GenerateRequest{
		Model:  req.Model,
		Prompt: promptText,
		System: req.System, // System prompt can be overridden or use default
	}

	// 3. Call LLM (Non-streaming for Analysis as JSON needs to be complete to parse)
	// We might need a separate "Rephrase" simple method if we want pure streaming speed without JSON
	rawResp, err := a.client.Generate(ctx, llmReq)
	if err != nil {
		return nil, err
	}

	// 4. Parse JSON Response
	parsedResp := ParseStructuredResponse(rawResp.Text)
	
	// If parsing failed (fallback to raw text), we still return it but Analysis will be nil
	return parsedResp, nil
}

// Rephrase keeps the original simple behavior (optional, or we can deprecate)
// For now, let's keep it simple: It just passes through, but ideally we use RephraseAndAnalyze for everything now.
func (a *PhraseSenseiAgent) Rephrase(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error) {
	// If the user of this method expects simple text, we should probably stick to simple prompting?
	// OR we use the new logic and just return the text part.
	// Let's use RephraseAndAnalyze but return only the text part to maintain compatibility where needed.
	
	// However, for TUI streaming, JSON parsing acts against streaming.
	// If we want real-time streaming characters, we can't wait for valid JSON.
	// CHOICE: For "Analysis" feature, we sacrifice some "first-byte latency" for "intelligence".
	// OR we stream the raw JSON and parse it on the client? (Complex)
	// Let's implement non-streaming for now as per the plan "The Analyst Agent" strategy 
	// mentions "single pass preferred".
	
	return a.RephraseAndAnalyze(ctx, req)
}

// RephraseStream - Warning: Streaming JSON is tricky.
// For now, we will NOT stream the analysis in real-time chunks that are parseable.
// We will accumulate the stream and parse at the end, OR disable streaming for Analysis mode.
// Given the requirement is to "Analyze", let's use the blocking Generate for the TUI's "Send" action 
// which is already structured as a command returning a Msg.
func (a *PhraseSenseiAgent) RephraseStream(ctx context.Context, req *GenerateRequest, fn func(*GenerateResponse) error) error {
	// For streaming, we might fallback to raw text OR try to stream the raw JSON tokens.
	// The current Router expects streaming.
	
	// If we use the new Prompt, the LLM will output JSON.
	// Streaming JSON characters to the user looks bad: `{"rephrased": "H...`
	
	// Strategy:
	// 1. If we want analysis, we likely can't do "character-by-character" update of the *rephrased* text 
	//    easily without a complex stream parser.
	// 2. Compromise: For the "Analyst" feature, we wait for the full response (Generate), 
	//    then pretend to stream it or just return it.
	
	// Let's wrap RephraseAndAnalyze and call the callback once at the end.
	resp, err := a.RephraseAndAnalyze(ctx, req)
	if err != nil {
		return err
	}
	
	return fn(resp)
}