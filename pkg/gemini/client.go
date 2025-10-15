package gemini

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Hardcoded API key for Console AI
const DefaultAPIKey = "AIzaSyC-gNO6yZPjN1XgS0k6ncidRMPeoQ72Z9U"

// NewClient creates and configures a new Gemini client.
// Uses hardcoded API key if none provided, defaults to gemini-2.5-flash model.
func NewClient(apiKey, modelName string) (*genai.GenerativeModel, error) {
	// Use hardcoded API key if none provided
	if apiKey == "" {
		apiKey = DefaultAPIKey
	}

	// Use latest model as default
	if modelName == "" {
		modelName = "gemini-2.5-flash"
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	model := client.GenerativeModel(modelName)
	model.Tools = defineTools()

	model.SafetySettings = []*genai.SafetySetting{
		{Category: genai.HarmCategoryHarassment, Threshold: genai.HarmBlockMediumAndAbove},
		{Category: genai.HarmCategoryHateSpeech, Threshold: genai.HarmBlockMediumAndAbove},
		{Category: genai.HarmCategorySexuallyExplicit, Threshold: genai.HarmBlockMediumAndAbove},
		{Category: genai.HarmCategoryDangerousContent, Threshold: genai.HarmBlockMediumAndAbove},
	}

	return model, nil
}
