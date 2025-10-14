package gemini

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

var systemPrompt string

func init() {
	// Load .env file. Errors are not fatal, as env vars can be set manually.
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it")
	}

	promptBytes, err := os.ReadFile("system_prompt.txt")
	if err != nil {
		log.Fatalf("Failed to read system prompt file: %v", err)
	}
	systemPrompt = string(promptBytes)
}

// NewClient creates a new Gemini client with a tool for executing shell commands.
func NewClient() (*genai.GenerativeModel, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// keep it gemini-2.5-flash for now don't change it
	model := client.GenerativeModel("gemini-2.5-flash")

	shellTool := &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{
			{
				Name:        "execute_shell_command",
				Description: "Executes a shell command on the user's machine.",
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"command": {
							Type:        genai.TypeString,
							Description: "The command to execute.",
						},
					},
					Required: []string{"command"},
				},
			},
		},
	}
	model.Tools = []*genai.Tool{shellTool}
	return model, nil
}

// ContinueConversation sends the conversation history to Gemini, handles tool calls, and gets a response.
func ContinueConversation(gemini *genai.GenerativeModel, history []string, input string, stepCallback func(title, content string)) (string, error) {
	ctx := context.Background()
	cs := gemini.StartChat()

	// Reconstruct chat history for the model
	for _, line := range history {
		if strings.HasPrefix(line, "User: ") {
			cs.History = append(cs.History, &genai.Content{Parts: []genai.Part{genai.Text(strings.TrimPrefix(line, "User: "))}, Role: "user"})
		} else if strings.HasPrefix(line, "Assistant: ") {
			cs.History = append(cs.History, &genai.Content{Parts: []genai.Part{genai.Text(strings.TrimPrefix(line, "Assistant: "))}, Role: "model"})
		}
	}

	// Prepend the system prompt to the user's first message in a session
	fullPrompt := input
	if len(history) == 0 {
		projectAnalysis, err := AnalyzeProject()
		if err != nil {
			log.Printf("Failed to analyze project: %v", err)
		}
		fullPrompt = systemPrompt + "\n\nProject Context:\n" + projectAnalysis + "\n\nUser: " + input
	}

	iter := cs.SendMessageStream(ctx, genai.Text(fullPrompt))
	var responseBuilder strings.Builder

	for {
		resp, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read stream: %w", err)
		}

		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			part := resp.Candidates[0].Content.Parts[0]
			if text, ok := part.(genai.Text); ok {
				chunk := string(text)
				responseBuilder.WriteString(chunk)
				stepCallback("Thinking...", chunk)
			} else if fc, ok := part.(genai.FunctionCall); ok {
				stepCallback("Function Call", fmt.Sprintf("Executing command: %s", fc.Args["command"]))
				// Handle function call
				command, ok := fc.Args["command"].(string)
				if !ok {
					return "", fmt.Errorf("invalid 'command' argument")
				}

				// Execute the command
				cmd := exec.Command("cmd", "/C", command)
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("Error executing command: %v", err)
				}

				// Send the result back to the model
				errStr := ""
				if err != nil {
					errStr = err.Error()
				}
				fr := genai.FunctionResponse{
					Name: "execute_shell_command",
					Response: map[string]interface{}{
						"output": string(output),
						"error":  errStr,
					},
				}

				// Send the function response back to the model in the same stream.
				iter = cs.SendMessageStream(ctx, fr)
			}
		}
	}

	return responseBuilder.String(), nil
}

// AnalyzeProject analyzes the project structure and files.
func AnalyzeProject() (string, error) {
	var projectStructure strings.Builder
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			projectStructure.WriteString(fmt.Sprintf("- %s\n", path))
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	return projectStructure.String(), nil
}