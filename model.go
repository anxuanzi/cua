package cua

import (
	"context"

	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

// createGeminiModel creates a Gemini model using the ADK gemini package.
func createGeminiModel(ctx context.Context, modelName string, cfg *genai.ClientConfig) (model.LLM, error) {
	return gemini.NewModel(ctx, modelName, cfg)
}
