package ollama

import (
	"context"
	"strings"
	"time"

	"github.com/ollama/ollama/api"
)

type ollamaGenerator struct {
	client *api.Client
	model  string
}

func NewOllamaGenerator(model string) (*ollamaGenerator, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, err
	}
	return &ollamaGenerator{client: client, model: model}, nil
}

func (g *ollamaGenerator) Generate(prompt string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	req := &api.GenerateRequest{
		Model:  g.model,
		Prompt: prompt,
	}

	var sb strings.Builder
	err := g.client.Generate(ctx, req, func(rsp api.GenerateResponse) error {
		sb.WriteString(rsp.Response)
		return nil
	})
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(sb.String()), nil
}
