package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"console-ai/pkg/config"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
)

const (
	// maxLoopIterations sets a hard limit on the number of tool-call cycles
	// to prevent infinite loops.
	maxLoopIterations = 15

	// conversationTimeout is the maximum duration for the entire conversation flow.
	conversationTimeout = 2 * time.Minute
)

// ContinueConversation handles the core logic of the AI's turn-based conversation.
// It sends the user's input to the Gemini model, processes tool calls, and streams
// the final text response back to the user interface.
func ContinueConversation(model *genai.GenerativeModel, history []string, input string, humorLevel int, cfg *config.Config, stepCallback func(title, content string)) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conversationTimeout)
	defer cancel()

	cs := model.StartChat()
	cs.History = buildHistory(history)

	if len(history) == 0 {
		toolDefinitions := generateToolDefinitions()
		dynamicPrompt := fmt.Sprintf(systemPrompt, toolDefinitions)
		dynamicPrompt += fmt.Sprintf("\n\nHumor Level: %d%%", humorLevel)
		model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(dynamicPrompt)}}
	}

	stepCallback("Thinking...", "")

	iter := cs.SendMessageStream(ctx, genai.Text(input))

	var responseBuilder strings.Builder
	var lastTextChunk string
	var hasResponded bool

	toolExecutor := NewToolExecutor(cfg)

	for i := 0; i < maxLoopIterations; i++ {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("stream error: %w", err)
		}

		if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			continue
		}

		for _, part := range resp.Candidates[0].Content.Parts {
			switch p := part.(type) {
			case genai.Text:
				textChunk := string(p)
				responseBuilder.WriteString(textChunk)
				if textChunk != lastTextChunk {
					stepCallback("Response", textChunk)
					lastTextChunk = textChunk
				}
				hasResponded = true

			case genai.FunctionCall:
				// Construct a more detailed message including function name and arguments
				argsJson, _ := json.Marshal(p.Args) // Safely marshal args to JSON
				stepCallback("Tool Call", fmt.Sprintf("\nExecuting: %s with args: %s", p.Name, string(argsJson)))
				output, err := toolExecutor.Execute(p)
				if err != nil {
					stepCallback("Tool Error", err.Error())
				}
				stepCallback("Tool Output", output)

				iter = cs.SendMessageStream(ctx, genai.FunctionResponse{
					Name:     p.Name,
					Response: map[string]interface{}{"output": output},
				})
			}
		}
	}
	// If the model finishes without generating a text response, provide a default message.
	if !hasResponded {
		return "The model finished its work without providing a direct response.", nil
	}

	return responseBuilder.String(), nil
}

// buildHistory reconstructs the conversation history from a simple string slice.
func buildHistory(history []string) []*genai.Content {
	if len(history) == 0 {
		return nil
	}

	var contents []*genai.Content
	for i := 0; i < len(history); i += 2 {
		userMessage := history[i]
		modelMessage := ""
		if i+1 < len(history) {
			modelMessage = history[i+1]
		}
		contents = append(contents, &genai.Content{Parts: []genai.Part{genai.Text(userMessage)}, Role: "user"})
		contents = append(contents, &genai.Content{Parts: []genai.Part{genai.Text(modelMessage)}, Role: "model"})
	}
	return contents
}
