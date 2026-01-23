package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/revenium/revenium-middleware-google-go/revenium"
	"google.golang.org/genai"
)

func main() {
	fmt.Println("=== Revenium Middleware - Google Gemini Vision Example ===")
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
		"taskType":       "vision-analysis",
		"agent":          "gemini-vision-example",
	}
	ctx = revenium.WithUsageMetadata(ctx, metadata)

	// Download and encode an image for Gemini Vision
	// Using a publicly available image URL for testing
	imageURL := "https://upload.wikimedia.org/wikipedia/commons/thumb/3/3a/Cat03.jpg/1200px-Cat03.jpg"

	fmt.Println("Downloading image for analysis...")
	imageData, mediaType, err := downloadImage(imageURL)
	if err != nil {
		log.Fatalf("Failed to download image: %v", err)
	}
	fmt.Printf("Image downloaded: %d bytes, type: %s\n\n", len(imageData), mediaType)

	// Create content with vision data
	fmt.Println("Analyzing image with Gemini Vision...")
	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: "What is in this image? Describe it in detail."},
				{
					InlineData: &genai.Blob{
						MIMEType: mediaType,
						Data:     imageData,
					},
				},
			},
		},
	}

	// Generate content with vision
	resp, err := client.Models().GenerateContent(
		ctx,
		"gemini-2.0-flash", // Gemini model with vision capabilities
		contents,
		nil,
	)
	if err != nil {
		log.Fatalf("Vision request failed: %v", err)
	}

	// Display response
	fmt.Println("\nVision Analysis:")
	fmt.Println(strings.Repeat("─", 45))

	if resp != nil && resp.Candidates != nil && len(resp.Candidates) > 0 {
		candidate := resp.Candidates[0]
		if candidate.Content != nil && candidate.Content.Parts != nil {
			for _, part := range candidate.Content.Parts {
				if part.Text != "" {
					fmt.Println(part.Text)
				}
			}
		}
	}

	fmt.Println(strings.Repeat("─", 45))
	fmt.Println()

	// Display usage stats
	if resp.UsageMetadata != nil {
		fmt.Printf("Model: %s\n", "gemini-2.0-flash")
		fmt.Printf("Input Tokens:  %d\n", resp.UsageMetadata.PromptTokenCount)
		fmt.Printf("Output Tokens: %d\n", resp.UsageMetadata.CandidatesTokenCount)
		fmt.Printf("Total Tokens:  %d\n", resp.UsageMetadata.TotalTokenCount)
	}

	// Wait for metering to complete
	client.Flush()

	fmt.Println("\n Vision analysis complete and metering data sent to Revenium!")
	fmt.Println("The metering payload includes hasVisionContent: true")
	fmt.Println("Check your Revenium dashboard for the metered usage.")
}

// downloadImage downloads an image from URL and returns the data and media type
func downloadImage(url string) ([]byte, string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, "", fmt.Errorf("failed to fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read image: %w", err)
	}

	// Get media type from Content-Type header
	mediaType := resp.Header.Get("Content-Type")
	if mediaType == "" {
		mediaType = "image/jpeg" // Default to JPEG
	}

	return data, mediaType, nil
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
