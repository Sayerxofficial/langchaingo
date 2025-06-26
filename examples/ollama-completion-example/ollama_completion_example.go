package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sayerxofficial/langchaingo/llms"
	"github.com/sayerxofficial/langchaingo/llms/ollama"
	"github.com/sayerxofficial/langchaingo/llms/streaming"
)

func main() {
	llm, err := ollama.New(ollama.WithModel("llama2"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()
	completion, err := llms.GenerateFromSinglePrompt(
		ctx,
		llm,
		"Human: Who was the first man to walk on the moon?\nAssistant:",
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			fmt.Println(chunk.String())
			return nil
		}),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = completion
}
