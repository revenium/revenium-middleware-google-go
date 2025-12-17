// Getting Started with Revenium Google AI Middleware
//
// This is the simplest example to verify your setup is working.
// Required environment variables:
// - GOOGLE_API_KEY
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

	fmt.Println("Testing Google AI with Revenium tracking...")

	// Create context with usage metadata
	// All supported metadata fields shown below (uncomment as needed)
	ctx := context.Background()
	metadata := map[string]any{
		// Required/Common fields
		"organizationId": "org-getting-started",
		"productId":      "product-getting-started",
		"taskType":       "text-generation",

		// Optional: Subscription and agent tracking
		// "subscriptionId": "sub-premium-tier",
		// "agent":          "my-agent-name",

		// Optional: Distributed tracing
		// "traceId": "trace-abc123-def456",

		// Optional: Quality scoring (0.0-1.0 scale)
		// "responseQualityScore": 0.95,

		// Optional: Subscriber details (for user attribution)
		// "subscriber": map[string]any{
		// 	"id":    "user-123",
		// 	"email": "user@example.com",
		// 	"credential": map[string]any{
		// 		"name":  "API Key Name",
		// 		"value": "key-identifier",
		// 	},
		// },
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
