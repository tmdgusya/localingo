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
    -   Outputs **Structured Data (JSON)**, ensuring every conversation yields actionable metadata.
    -   **Meta-Analysis:** Can digest multiple past interactions to find high-level patterns.

2.  **The Memory (PostgreSQL & SRS):**
    -   Stores conversation history and **Analysis Metadata**.
    -   **Spaced Repetition System (SRS):** Implements the **SM-2 algorithm** to track the "forgetting curve" of specific user mistakes.
    -   Data is structured as `CorrectionPair` (Original vs. Corrected) for targeted training.

3.  **The Interface (TUI):**
    -   A distraction-free, terminal-based environment for focused learning.
    -   **Command Center:** Real-time hints and discovery via the `/` command bar.
    -   **Multi-Modal Views:** Seamlessly switches between Chat, History, Analysis Report, Lessons, and Review Quizzes.

## 3. Current Implementation Status (The Complete Learning Cycle)

### ✅ Completed
-   **Prompt Management:** Externalized templates for flexible pedagogy.
-   **Structured Agent:** Real-time tagging of Grammar, Vocabulary, and Naturalness.
-   **Pattern Analyst (`/report`):** Quantitative visualization of recurring error categories.
-   **Private Tutor (`/lesson`):** Qualitative meta-analysis of past mistakes to provide "Native Mindset" coaching.
-   **Fossil Breaker (`/review`):** Active recall training using the SM-2 SRS algorithm to break bad habits.
-   **UX & Guidance:** Real-time command hints and auto-completion logic.
-   **Observability:** Robust file logging (`debug.log`) for system health monitoring.

---
*Localingo is not just a translation tool; it is a personalized language observatory and a scientific laboratory for habit-breaking.*

---
*Localingo is not just a translation tool; it is a personalized language observatory.*
