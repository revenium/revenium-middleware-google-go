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
		"organizationId": "org-basic-example",
		"productId":      "product-basic",
		"taskType":       "text-generation",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	model := "gemini-2.0-flash-exp"
	fmt.Printf("Generating content with model: %s\n", model)

	resp, err := client.Models().GenerateContent(
		ctx,
		model,
		genai.Text("Explain how AI works in 3 sentences"),
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	fmt.Println("\n=== Response ===")
	fmt.Println(resp.Text())

	if resp.UsageMetadata != nil {
		fmt.Println("\n=== Usage Metadata ===")
		fmt.Printf("Prompt tokens: %d\n", resp.UsageMetadata.PromptTokenCount)
		fmt.Printf("Candidates tokens: %d\n", resp.UsageMetadata.CandidatesTokenCount)
		fmt.Printf("Total tokens: %d\n", resp.UsageMetadata.TotalTokenCount)
		if resp.UsageMetadata.CachedContentTokenCount > 0 {
			fmt.Printf("Cached tokens: %d\n", resp.UsageMetadata.CachedContentTokenCount)
		}
		if resp.UsageMetadata.ThoughtsTokenCount > 0 {
			fmt.Printf("Thoughts tokens: %d\n", resp.UsageMetadata.ThoughtsTokenCount)
		}
	}

	fmt.Println("\nMetering data sent to Revenium in the background")
}
