package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/revenium/revenium-middleware-google-go/revenium"
	"google.golang.org/genai"
)

func main() {
	fmt.Println("=== Revenium Middleware - Google Veo Video Generation Example ===")
	fmt.Println()

	// Initialize the middleware
	// Required env vars:
	// - GOOGLE_API_KEY: Your Google API key (for Google AI)
	// - REVENIUM_METERING_API_KEY: Your Revenium metering API key (hak_...)
	// Optional:
	// - REVENIUM_METERING_BASE_URL: Override Revenium API URL (default: https://api.revenium.ai)
	

	if err := revenium.Initialize(); err != nil {
		log.Fatalf("Failed to initialize middleware: %v", err)
	}

	// Get the client
	client, err := revenium.GetClient()
	if err != nil {
		log.Fatalf("Failed to get client: %v", err)
	}
	defer client.Close()

	// Create context with custom metadata for billing/tracking
	ctx := context.Background()
	metadata := map[string]interface{}{
		"organizationId": "org-123",
		"productId":      "prod-456",
		"taskType":       "video-generation",
		"agent":          "veo-example",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Generate a video with Veo
	fmt.Println("Starting video generation with Veo...")
	prompt := "A timelapse of a flower blooming in a garden, soft natural lighting"

	// Configure video generation
	config := &genai.GenerateVideosConfig{
		NumberOfVideos: 1,
		AspectRatio:    "16:9",
	}

	// Start video generation (asynchronous operation)
	// Note: Veo models include: veo-2.0-generate-001
	operation, err := client.Videos().GenerateVideos(
		ctx,
		"veo-2.0-generate-001",
		prompt,
		nil, // No reference image
		config,
	)
	if err != nil {
		log.Fatalf("Video generation failed to start: %v", err)
	}

	fmt.Printf("Video generation started! Operation: %s\n", operation.Name)
	fmt.Println()

	// Wait for completion (this will poll and meter when done)
	fmt.Println("Waiting for video generation to complete...")
	fmt.Println("(This may take several minutes)")
	fmt.Println()

	resp, err := client.Videos().WaitForVideoGeneration(
		ctx,
		operation,
		"veo-2.0-generate-001",
		10*time.Second, // Poll every 10 seconds
		5*time.Minute,  // Timeout after 5 minutes
	)
	if err != nil {
		log.Fatalf("Video generation failed: %v", err)
	}

	// Display results
	fmt.Println()
	fmt.Println("Video Generation Results:")
	fmt.Println(strings.Repeat("─", 45))

	if resp != nil && resp.GeneratedVideos != nil {
		fmt.Printf("Videos generated: %d\n", len(resp.GeneratedVideos))
		for i, video := range resp.GeneratedVideos {
			if video.Video != nil && video.Video.URI != "" {
				fmt.Printf("Video %d: %s\n", i+1, video.Video.URI)
			}
		}
	}

	// Show RAI filtering info if any
	if resp != nil && resp.RAIMediaFilteredCount > 0 {
		fmt.Printf("\nNote: %d videos were filtered by RAI\n", resp.RAIMediaFilteredCount)
		if len(resp.RAIMediaFilteredReasons) > 0 {
			fmt.Printf("Reasons: %v\n", resp.RAIMediaFilteredReasons)
		}
	}

	fmt.Println(strings.Repeat("─", 45))
	fmt.Println()

	// Wait for metering to complete
	client.Flush()

	fmt.Println("Video generation complete and metering data sent to Revenium!")
	fmt.Println("The metering payload includes:")
	fmt.Println("  - operationType: VIDEO")
	fmt.Println("  - actualVideoCount: number of videos returned")
	fmt.Println("  - requestedVideoCount: number requested")
	fmt.Println("Check your Revenium dashboard for the metered usage.")
}

func init() {
	// Validate required env vars
	if os.Getenv("GOOGLE_API_KEY") == "" {
		log.Fatal("GOOGLE_API_KEY environment variable is required")
	}
	if os.Getenv("REVENIUM_METERING_API_KEY") == "" {
		log.Fatal("REVENIUM_METERING_API_KEY environment variable is required")
	}
}
