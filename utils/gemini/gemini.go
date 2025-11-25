package gemini

import (
	"context"
	"log"

	"github.com/kacperborowieckb/gen-sql/utils/env"
	"google.golang.org/genai"
)

func NewConnection() (*genai.Client, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: env.GetString("GEMINI_API_KEY", ""),
	})

	if err != nil {
		return nil, err
	}

	log.Printf("new gemini client set up")

	return client, nil
}
