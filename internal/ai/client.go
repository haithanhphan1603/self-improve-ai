package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

var client *openai.Client

func InitClient() error {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}
	client = openai.NewClient(apiKey)
	return nil
}

type FeedbackResponse struct {
	Summary  string `json:"summary"`
	Feedback string `json:"feedback"`
}

func GetJournalFeedback(content string) (string, string, error) {
	if client == nil {
		return "", "", fmt.Errorf("AI client not initialized")
	}

	if strings.TrimSpace(content) == "" {
		return "", "", fmt.Errorf("journal content cannot be empty")
	}

	prompt := `You are a compassionate self-improvement coach AI.

Analyze the user's journal entry and provide:
1. A brief, encouraging summary (2-3 sentences)
2. Constructive feedback with actionable insights

Return ONLY valid JSON in this exact format:
{
  "summary": "Brief summary here",
  "feedback": "Constructive feedback with specific suggestions"
}

User's Journal Entry:
` + content

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       openai.GPT4oMini, // More cost-effective than GPT-4
		Temperature: 0.7,              // Balanced creativity
		MaxTokens:   500,              // Limit response length
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a supportive self-improvement coach. Always respond with valid JSON only.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	})
	if err != nil {
		return "", "", fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", "", fmt.Errorf("no response from OpenAI")
	}

	raw := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Clean up potential markdown formatting
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var parsed FeedbackResponse
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		// Improved fallback with error details
		return "Unable to parse AI response",
			fmt.Sprintf("AI provided feedback but in unexpected format: %s", raw),
			nil
	}

	// Validate response quality
	if parsed.Summary == "" || parsed.Feedback == "" {
		return "", "", fmt.Errorf("AI returned incomplete response")
	}

	return parsed.Summary, parsed.Feedback, nil
}
