package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/revenium/revenium-middleware-google-go/revenium"
	"google.golang.org/genai"
)

func main() {
	fmt.Println("=== Revenium Middleware - Google Imagen Example ===")
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
		"taskType":       "image-generation",
		"agent":          "imagen-example",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Generate an image with Imagen
	fmt.Println("Generating image with Imagen...")
	prompt := "A serene mountain landscape at sunset with a calm lake reflecting the colorful sky"

	// Configure image generation
	config := &genai.GenerateImagesConfig{
		NumberOfImages: 1,
		AspectRatio:    "16:9",
	}

	// Generate the image
	// Note: Imagen models include: imagen-3.0-generate-001, imagen-3.0-fast-generate-001
	resp, err := client.Images().GenerateImages(
		ctx,
		"imagen-3.0-generate-001",
		prompt,
		config,
	)
	if err != nil {
		log.Fatalf("Image generation failed: %v", err)
	}

	// Display results
	fmt.Println()
	fmt.Println("Image Generation Results:")
	fmt.Println(strings.Repeat("─", 45))

	if resp != nil && resp.GeneratedImages != nil {
		fmt.Printf("Images generated: %d\n", len(resp.GeneratedImages))
		for i, img := range resp.GeneratedImages {
			if img.Image != nil && img.Image.ImageBytes != nil {
				// Show first 50 chars of base64 as preview
				b64 := base64.StdEncoding.EncodeToString(img.Image.ImageBytes)
				preview := b64
				if len(preview) > 50 {
					preview = preview[:50] + "..."
				}
				fmt.Printf("Image %d: %d bytes (base64 preview: %s)\n", i+1, len(img.Image.ImageBytes), preview)
			}
		}
	}

	fmt.Println(strings.Repeat("─", 45))
	fmt.Println()

	// Wait for metering to complete
	client.Flush()

	fmt.Println("Image generation complete and metering data sent to Revenium!")
	fmt.Println("The metering payload includes:")
	fmt.Println("  - operationType: IMAGE")
	fmt.Println("  - actualImageCount: number of images returned")
	fmt.Println("  - requestedImageCount: number requested")
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
