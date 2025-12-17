// Getting Started with Revenium Vertex AI Middleware
//
// This is the simplest example to verify your Vertex AI setup is working.
// Required environment variables:
// - GOOGLE_CLOUD_PROJECT
// - GOOGLE_CLOUD_LOCATION
// - GOOGLE_APPLICATION_CREDENTIALS
// - REVENIUM_METERING_API_KEY

package main

import (
	"context"
	"fmt"
	"log"

	"github.com/revenium/revenium-middleware-google-go/revenium"
	"google.golang.org/genai"
)

func main() {
	// Initialize middleware (automatically uses environment variables)
	if err := revenium.Initialize(); err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	client, err := revenium.GetClient()
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}
	defer client.Close()

	fmt.Println("Testing Vertex AI with Revenium tracking...")

	// Simple chat completion
	ctx := context.Background()
	metadata := map[string]any{
		"organizationId": "org-vertex-getting-started",
		"productId":      "product-vertex-getting-started",
		"taskType":       "vertex-text-generation",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)
	resp, err := client.Models().GenerateContent(
		ctx,
		"gemini-2.0-flash-exp",
		genai.Text("Please verify you are ready to assist me."),
		nil,
	)

	if err != nil {
		log.Fatalf("Failed to generate content: %v", err)
	}

	fmt.Println("Response:", resp.Text())

	if resp.UsageMetadata != nil {
		fmt.Println("\nUsage:")
		fmt.Printf("  Total tokens: %d\n", resp.UsageMetadata.TotalTokenCount)
		fmt.Printf("  Prompt tokens: %d\n", resp.UsageMetadata.PromptTokenCount)
		fmt.Printf("  Candidates tokens: %d\n", resp.UsageMetadata.CandidatesTokenCount)
	}

	fmt.Println("\nTracking successful! Check your Revenium dashboard.")
}
