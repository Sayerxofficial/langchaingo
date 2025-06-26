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
	llm, err := openai.New()
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What would be a good company name a company that makes colorful socks?"),
	}

	if r, err := llm.GenerateContent(
		ctx,
		content,
		llms.WithMaxTokens(1024),
		llms.WithStreamingFunc(func(_ context.Context, chunk streaming.Chunk) error {
			fmt.Println(chunk.String())
			return nil
		}),
	); err != nil {
		log.Fatal(err)
	} else {
		fmt.Println(r.Choices[0].Content)
	}
}
