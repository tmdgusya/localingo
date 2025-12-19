package agent

type PhraseSenseiAgent struct {
	client LLMClient
}

func NewPhraseSenseiAgent(client LLMClient) *PhraseSenseiAgent {
	return &PhraseSenseiAgent{
		client: client,
	}
}
