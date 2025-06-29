package main

import (
	"context"
	"fmt"
	"log"

	"github.com/sayerxofficial/langchaingo/llms"
	"github.com/sayerxofficial/langchaingo/llms/openai"
	"github.com/sayerxofficial/langchaingo/llms/streaming"
)

func main() {
	llm, err := openai.New(openai.WithModel("gpt-4o"))
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a ocr assistant."),
		llms.TextParts(llms.ChatMessageTypeHuman, "which image is this?"),
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.ImageURLPart("https://langchaingo/blob/main/docs/static/img/parrot-icon.png?raw=true")},
		},
	}

	completion, err := llm.GenerateContent(
		ctx,
		content,
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
