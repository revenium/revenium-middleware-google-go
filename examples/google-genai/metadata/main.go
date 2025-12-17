package main

import (
	"context"
	"fmt"
	"log"

	"github.com/revenium/revenium-middleware-google-go/revenium"
	"google.golang.org/genai"
)

func main() {
	ctx := context.Background()

	// Initialize Revenium middleware
	if err := revenium.Initialize(); err != nil {
		log.Fatalf("Failed to initialize Revenium: %v", err)
	}

	// Get the Revenium client
	client, err := revenium.GetClient()
	if err != nil {
		log.Fatalf("Failed to get Revenium client: %v", err)
	}
	defer client.Close()

	fmt.Println("Testing Google AI with complete metadata tracking...")
	fmt.Println()

	// Create comprehensive metadata with all supported fields
	metadata := map[string]interface{}{
		// Organization and product tracking
		"organizationId": "org-acme-corp",
		"productId":      "product-ai-assistant",
		"subscriptionId": "sub-premium-tier",

		// Task and agent tracking
		"taskType": "customer-support",
		"agent":    "support-bot-v2",

		// Tracing and correlation
		"traceId": "trace-abc123-def456",

		// Quality metrics
		"responseQualityScore": 95.5,

		// Subscriber information (complete object)
		"subscriber": map[string]interface{}{
			"id":    "user-john-doe-789",
			"email": "john.doe@example.com",
			"credential": map[string]interface{}{
				"name":  "Production API Key",
				"value": "pk-prod-xyz789",
			},
		},
	}

	// Add metadata to context
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Configure the generation with temperature
	temperature := float32(0.7)
	config := &genai.GenerateContentConfig{
		Temperature: &temperature, // This will be automatically captured
	}

	// Generate content with metadata tracking
	resp, err := client.Models().GenerateContent(
		ctx,
		"gemini-2.0-flash-exp",
		genai.Text("Explain the importance of metadata in AI systems in one paragraph."),
		config,
	)
	if err != nil {
		log.Fatalf("Error generating content: %v", err)
	}

	fmt.Println("Response:", resp.Text())
	fmt.Println()

	// Display usage information
	if resp.UsageMetadata != nil {
		fmt.Println("Usage:")
		fmt.Printf("  Total tokens: %d\n", resp.UsageMetadata.TotalTokenCount)
		fmt.Printf("  Prompt tokens: %d\n", resp.UsageMetadata.PromptTokenCount)
		fmt.Printf("  Candidates tokens: %d\n", resp.UsageMetadata.CandidatesTokenCount)
		if resp.UsageMetadata.CachedContentTokenCount > 0 {
			fmt.Printf("  Cached tokens: %d\n", resp.UsageMetadata.CachedContentTokenCount)
		}
		fmt.Println()
	}

	fmt.Println("Tracking successful! Check your Revenium dashboard for complete metadata.")
	fmt.Println()
	fmt.Println("All metadata fields sent")
}
