package gemini

import (
	"context"
	"fmt"
	"strings"
	"time"

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
func ContinueConversation(model *genai.GenerativeModel, history []string, input string, stepCallback func(title, content string)) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), conversationTimeout)
	defer cancel()

	// Start a new chat session and reconstruct the history.
	cs := model.StartChat()
	cs.History = buildHistory(history)

	// Use the first message as the system prompt if it's the start of the conversation.
	if len(history) == 0 {
		model.SystemInstruction = &genai.Content{Parts: []genai.Part{genai.Text(systemPrompt)}}
	}

	stepCallback("Thinking...", "")

	// Send the user's message and start the processing loop.
	iter := cs.SendMessageStream(ctx, genai.Text(input))

	var responseBuilder strings.Builder
	var lastTextChunk string
	var hasResponded bool

	for i := 0; i < maxLoopIterations; i++ {
		resp, err := iter.Next()
		if err == iterator.Done {
			break // Normal end of conversation turn.
		}
		if err != nil {
			return "", fmt.Errorf("stream error: %w", err)
		}

		if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			continue // Skip empty responses.
		}

		// Process each part of the response (text or function call).
		for _, part := range resp.Candidates[0].Content.Parts {
			switch p := part.(type) {
			case genai.Text:
				// Append text to the response and notify the UI.
				textChunk := string(p)
				responseBuilder.WriteString(textChunk)
				// To avoid spamming the UI, only send updates when the text changes.
				if textChunk != lastTextChunk {
					stepCallback("Response", textChunk)
					lastTextChunk = textChunk
				}
				hasResponded = true

			case genai.FunctionCall:
				// Execute the requested tool and send the result back to the model.
				stepCallback("Tool Call", fmt.Sprintf("Executing: %s", p.Name))
				output, err := executeTool(p)
				if err != nil {
					stepCallback("Tool Error", err.Error())
				}
				stepCallback("Tool Output", output)

				// Send the tool's output back to the model to continue the loop.
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
