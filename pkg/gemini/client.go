package gemini

import (
	"context"
	"fmt"
	"os"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// NewClient creates and configures a new Gemini client.
// It now loads the API key directly from the constants file.
func NewClient() (*genai.GenerativeModel, error) {
	apiKey := geminiAPIKey // Using the constant from constants.go
	if apiKey == "" {
		// This check remains as a safeguard, though it should always be present.
		return nil, fmt.Errorf("gemini API key is not set in constants.go")
	}

	// Check for a user-provided model name from environment variables,
	// otherwise default to "gemini-1.5-flash".
	modelName := os.Getenv("GEMINI_MODEL")
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	// Configure the generative model with the selected model name and tools
	model := client.GenerativeModel(modelName)
	model.Tools = defineTools() // Assumes defineTools() is in the same package

	// Set safety settings to block harmful content
	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockMediumAndAbove},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockMediumAndAbove},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockMediumAndAbove},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockMediumAndAbove},
	}

	return model, nil
}
