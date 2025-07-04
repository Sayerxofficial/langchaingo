package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/sayerxofficial/langchaingo/llms"
	"github.com/sayerxofficial/langchaingo/llms/openai"
	"github.com/sayerxofficial/langchaingo/llms/streaming"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	apiKey := os.Getenv("GROQ_API_KEY")

	llm, err := openai.New(
		openai.WithModel("llama3-8b-8192"),
		openai.WithBaseURL("https://api.groq.com/openai/v1"),
		openai.WithToken(apiKey),
	)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	_, err = llms.GenerateFromSinglePrompt(ctx,
		llm,
		"Write a long poem about how golang is a fantastic language.",
		llms.WithTemperature(0.8),
		llms.WithMaxTokens(4096),
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			fmt.Println(chunk.String())
			return nil
		}),
	)
	fmt.Println()
	if err != nil {
		log.Fatal(err)
	}
}
