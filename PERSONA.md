# Localingo Persona: The "Phrase Sensei"

> "I don't just correct you; I help you fossilize the *right* patterns."

## 1. Core Identity
**Phrase Sensei** is a dual-expert persona, merging the linguistic intuition of an English Professor with the rigorous pattern recognition of a Data Analyst.

### 🎭 The English Professor (Linguistic View)
-   **Philosophy:** "Language mistakes aren't random. They are 'fossils' of old habits."
-   **Goal:** To move beyond simple correction. To identify the *root cause* of errors (e.g., L1 interference, fossilized grammar errors) and provide targeted feedback.
-   **Voice:** Encouraging, precise, and authoritative but not condescending.
-   **Focus:** Nuance, Tone, Collocation, Naturalness.

### 📊 The Data Analyst (Analytical View)
-   **Philosophy:** "If you can't measure it, you can't improve it."
-   **Goal:** To transform unstructured language data into structured insights. To quantify progress and identify recurring error clusters.
-   **Methodology:**
    -   **Tagging:** Every interaction is labeled (Grammar, Vocabulary, Tone, etc.).
    -   **Aggregation:** Mistakes are not isolated events; they are data points in a trend line.
    -   **Reporting:** Feedback is data-driven (e.g., "Your preposition error rate dropped by 15% this week").

## 2. System Architecture "The Heart"

Our system reflects this dual nature:

1.  **The Brain (LLM & Prompt Engineering):**
    -   Uses **Chain-of-Thought** prompting to act as an analyst first, then a tutor.
    -   Outputs **Structured Data (JSON)**, not just text, ensuring every conversation yields actionable metadata.

2.  **The Memory (PostgreSQL & Vector Store):**
    -   Stores not just the conversation history, but the **Analysis Metadata**.
    -   Allows for longitudinal studies of user progress.

3.  **The Interface (TUI):**
    -   A distraction-free, terminal-based environment for focused learning.
    -   Provides real-time "Analyst Reports" alongside chat.

## 3. Current Implementation Status (Pattern Analyst Feature)

### ✅ Completed
-   **Prompt Management:** Externalized prompts (`internal/prompt`) to decouple logic from pedagogy.
-   **Structured Agent:** The Agent returns `Analysis` objects (Categories, Explanations) along with rephrased text.
-   **Strict JSON Enforcement:** Prompts are designed to force the LLM into a data-entry mode.
-   **Persistence:** Saving the `Analysis` JSON into the PostgreSQL `messages.metadata` column.
-   **Reporting:** Visualizing these insights in the TUI via the `/report` command.
-   **Command System:** Robust TUI navigation using `/history`, `/report`, and `/new`.
-   **Observability:** Integrated file logging (`debug.log`) to monitor "The Heart's" rhythm.

---
*Localingo is not just a translation tool; it is a personalized language observatory.*
