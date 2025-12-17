package main

import (
	"context"
	"fmt"
	"log"

	"github.com/revenium/revenium-middleware-google-go/revenium"
	"google.golang.org/genai"
)

func main() {
	if err := revenium.Initialize(); err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	client, err := revenium.GetClient()
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	metadata := map[string]any{
		"organizationId": "org-vertex-streaming",
		"productId":      "product-vertex-streaming",
		"taskType":       "vertex-streaming-generation",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	fmt.Println("=== Streaming Response ===")

	stream := client.Models().GenerateContentStream(
		ctx,
		"gemini-2.0-flash-exp",
		genai.Text("Write a short story about a magic backpack in 3 paragraphs"),
		nil,
	)

	for chunk, err := range stream {
		if err != nil {
			log.Fatalf("Stream error: %v", err)
		}

		if chunk != nil && len(chunk.Candidates) > 0 {
			fmt.Print(chunk.Text())
		}
	}

	fmt.Println("\n\nStreaming complete. Metering data sent to Revenium in the background")
}
