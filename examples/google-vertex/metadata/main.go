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

	fmt.Println("Testing Vertex AI with complete metadata tracking...")
	fmt.Println()

	// Create comprehensive metadata with all supported fields
	metadata := map[string]interface{}{
		// Organization and product tracking
		"organizationId": "org-vertex-enterprise",
		"productId":      "product-vertex-assistant",
		"subscriptionId": "sub-enterprise-tier",

		// Task and agent tracking
		"taskType": "data-analysis",
		"agent":    "vertex-analytics-bot-v3",

		// Tracing and correlation
		"traceId": "trace-vertex-xyz789-abc123",

		// Quality metrics
		"responseQualityScore": 98.2,

		// Subscriber information (complete object)
		"subscriber": map[string]interface{}{
			"id":    "user-vertex-analyst-456",
			"email": "analyst@enterprise.com",
			"credential": map[string]interface{}{
				"name":  "Vertex Production Key",
				"value": "vk-prod-abc456",
			},
		},
	}

	// Add metadata to context
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Configure the generation with temperature
	temperature := float32(0.8)
	config := &genai.GenerateContentConfig{
		Temperature: &temperature, // This will be automatically captured
	}

	// Generate content with metadata tracking
	resp, err := client.Models().GenerateContent(
		ctx,
		"gemini-2.0-flash-exp",
		genai.Text("Explain how Vertex AI helps enterprises scale their AI workloads in one paragraph."),
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

	// Display metadata that was sent
	fmt.Println("Metadata sent to Revenium")
	fmt.Println()
	fmt.Println("Tracking successful! Check your Revenium dashboard for complete metadata.")
}
