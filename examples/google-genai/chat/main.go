package main

import (
	"context"
	"fmt"
	"log"

	"github.com/revenium/revenium-middleware-google-go/revenium"
	"google.golang.org/genai"
)

func main() {
	// Initialize the middleware
	if err := revenium.Initialize(); err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	// Get the client
	client, err := revenium.GetClient()
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}
	defer client.Close()

	// Create context with custom metadata
	ctx := context.Background()
	metadata := map[string]interface{}{
		"organizationId": "org-chat-example",
		"productId":      "product-chat",
		"taskType":       "multi-turn-chat",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Multi-turn conversation
	fmt.Println("=== Multi-turn Chat Example ===")

	// Turn 1: User asks about AI
	fmt.Println("User: What is artificial intelligence?")
	resp1, err := client.Models().GenerateContent(
		ctx,
		"gemini-2.0-flash-exp",
		genai.Text("What is artificial intelligence?"),
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}
	fmt.Printf("AI: %s\n\n", resp1.Text())

	// Turn 2: Follow-up question using conversation history
	fmt.Println("User: Can you give me an example?")

	// Build conversation history
	history := []*genai.Content{
		genai.NewContentFromText("What is artificial intelligence?", genai.RoleUser),
		genai.NewContentFromText(resp1.Text(), genai.RoleModel),
		genai.NewContentFromText("Can you give me an example?", genai.RoleUser),
	}

	resp2, err := client.Models().GenerateContent(
		ctx,
		"gemini-2.0-flash-exp",
		history,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}
	fmt.Printf("AI: %s\n\n", resp2.Text())

	// Turn 3: Another follow-up
	fmt.Println("User: How does it learn?")

	// Update conversation history
	history = append(history,
		genai.NewContentFromText(resp2.Text(), genai.RoleModel),
		genai.NewContentFromText("How does it learn?", genai.RoleUser),
	)

	resp3, err := client.Models().GenerateContent(
		ctx,
		"gemini-2.0-flash-exp",
		history,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}
	fmt.Printf("AI: %s\n\n", resp3.Text())

	// Print total usage
	fmt.Println("=== Total Usage ===")
	totalTokens := resp1.UsageMetadata.TotalTokenCount +
		resp2.UsageMetadata.TotalTokenCount +
		resp3.UsageMetadata.TotalTokenCount
	fmt.Printf("Total tokens across all turns: %d\n", totalTokens)
	fmt.Printf("Turn 1: %d tokens\n", resp1.UsageMetadata.TotalTokenCount)
	fmt.Printf("Turn 2: %d tokens\n", resp2.UsageMetadata.TotalTokenCount)
	fmt.Printf("Turn 3: %d tokens\n", resp3.UsageMetadata.TotalTokenCount)

	fmt.Println("\nAll metering data sent to Revenium")
}
