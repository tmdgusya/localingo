# Feature Implementation Plan: "Pattern Analyst"

## 1. Objective
Identify and report recurring patterns in the user's English mistakes based on the original input and the agent's rephrased output.

## 2. Strategy: "The Analyst Agent"
We will utilize the LLM not just for rephrasing, but for **Analysis & Tagging** in a single pass (or a specialized second pass if latency permits, but single pass is preferred for speed).

### Core Components
1.  **Prompts as Code (Configurable):** Move hardcoded prompts to external text/template files or a configuration structure.
2.  **Structured Output:** The Agent must return structured data (JSON) containing:
    *   `rephrased_text`: The polished sentence.
    *   `analysis`: A list of detected issues/improvements.
3.  **Storage:** Save this structured analysis into the existing `metadata` JSONB column in PostgreSQL.
4.  **Reporting:** A new TUI view to visualize aggregated stats.

## 3. Implementation Steps

### Phase 1: Configuration & Prompt Management
- [ ] Create `internal/prompt` package to load prompts from files/assets.
- [ ] Create default prompt templates for:
    -   `system_rephrase_and_analyze.txt`: Instructs the model to rephrase AND analyze errors.

### Phase 2: Agent Refactoring
- [ ] Update `GenerateResponse` struct to include `Analysis` data.
- [ ] Update `PhraseSenseiAgent` to parse JSON output from the LLM.
    -   *Challenge:* LLMs can be chatty. We need to enforce JSON format (using `format: "json"` in Ollama API if supported, or strict prompting).

### Phase 3: Persistence
- [ ] Update `HandleRephraseStream` in Router to extract analysis data from the Agent's response.
- [ ] Save this data into the `metadata` field when calling `CreateMessage`.

### Phase 4: TUI & Visualization
- [ ] Add `/report` command in TUI.
- [ ] Implement a simple aggregation query in `ChatHistoryRepository`.
    -   `GetErrorStats(ctx, days int) map[string]int`
- [ ] Render a simple bar chart or list in TUI showing top error types.

## 4. Prompt Design Draft
*(To be externalized)*

```text
You are an expert English Professor and Data Analyst.
Your task is to rephrase the user's input to be more natural and native-like.
CRITICALLY, you must also analyze what weak points the user has.

Return ONLY a JSON object in the following format:
{
    "rephrased": "The corrected sentence here.",
    "analysis": {
        "categories": ["Grammar", "Nuance", "Typo", "Idiom"],
        "explanation": "Brief explanation of why you changed it."
    }
}

User Input: {{.Input}}
```
