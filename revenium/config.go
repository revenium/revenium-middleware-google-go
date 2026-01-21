package revenium

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the Revenium middleware
type Config struct {
	GoogleAPIKey string

	ProjectID string
	Location  string

	ReveniumAPIKey  string
	ReveniumBaseURL string

	VertexDisabled bool
	Debug          bool

	// Prompt capture configuration (opt-in)
	CapturePrompts bool
}

// Option is a functional option for configuring Config
type Option func(*Config)

// WithGoogleAPIKey sets the Google API key
func WithGoogleAPIKey(key string) Option {
	return func(c *Config) {
		c.GoogleAPIKey = key
	}
}

// WithProjectID sets the Google Cloud Project ID for Vertex AI
func WithProjectID(projectID string) Option {
	return func(c *Config) {
		c.ProjectID = projectID
	}
}

// WithLocation sets the Google Cloud Location for Vertex AI
func WithLocation(location string) Option {
	return func(c *Config) {
		c.Location = location
	}
}

// WithReveniumAPIKey sets the Revenium API key
func WithReveniumAPIKey(key string) Option {
	return func(c *Config) {
		c.ReveniumAPIKey = key
	}
}

// WithReveniumBaseURL sets the Revenium base URL
func WithReveniumBaseURL(url string) Option {
	return func(c *Config) {
		c.ReveniumBaseURL = url
	}
}

// WithDebug enables or disables debug logging programmatically
func WithDebug(debug bool) Option {
	return func(c *Config) {
		c.Debug = debug
	}
}

// WithCapturePrompts enables or disables prompt capture for analytics
// When enabled, system prompts, input messages, and output responses are captured
// and sent to Revenium for analytics (with truncation at 50,000 characters)
func WithCapturePrompts(capture bool) Option {
	return func(c *Config) {
		c.CapturePrompts = capture
	}
}

// loadFromEnv loads configuration from environment variables and .env files
func (c *Config) loadFromEnv() error {
	// First, try to load .env files automatically
	c.loadEnvFiles()

	// Then load from environment variables (which may have been set by .env files)
	c.GoogleAPIKey = os.Getenv("GOOGLE_API_KEY")
	c.ProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	c.Location = os.Getenv("GOOGLE_CLOUD_LOCATION")

	c.ReveniumAPIKey = os.Getenv("REVENIUM_METERING_API_KEY")
	baseURL := os.Getenv("REVENIUM_METERING_BASE_URL")
	if baseURL == "" {
		baseURL = defaultReveniumBaseURL
	}
	c.ReveniumBaseURL = NormalizeReveniumBaseURL(baseURL)

	if os.Getenv("REVENIUM_VERTEX_DISABLE") == "1" || os.Getenv("REVENIUM_VERTEX_DISABLE") == "true" {
		c.VertexDisabled = true
	}

	c.Debug = os.Getenv("REVENIUM_DEBUG") == "true" || os.Getenv("REVENIUM_DEBUG") == "1"
	c.CapturePrompts = os.Getenv("REVENIUM_CAPTURE_PROMPTS") == "true" || os.Getenv("REVENIUM_CAPTURE_PROMPTS") == "1"

	// Initialize logger early so we can use it
	InitializeLogger()

	// Set global debug flag for logger
	SetGlobalDebug(c.Debug)

	Debug("Loading configuration from environment variables")

	return nil
}

// loadEnvFiles loads environment variables from .env files
func (c *Config) loadEnvFiles() {
	// Try to load .env files in order of preference
	envFiles := []string{
		".env.local", // Local overrides (highest priority)
		".env",       // Main env file
	}

	// Get current working directory
	cwd, err := os.Getwd()
	if err != nil {
		cwd = "."
	}

	// Try current directory and parent directories
	searchDirs := []string{
		cwd,
		filepath.Dir(cwd),
		filepath.Join(cwd, ".."),
	}

	for _, dir := range searchDirs {
		for _, envFile := range envFiles {
			envPath := filepath.Join(dir, envFile)

			// Check if file exists
			if _, err := os.Stat(envPath); err == nil {
				// Try to load the file
				_ = godotenv.Load(envPath)
			}
		}
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.ReveniumAPIKey == "" {
		return NewConfigError("REVENIUM_METERING_API_KEY is required", nil)
	}

	if !isValidAPIKeyFormat(c.ReveniumAPIKey) {
		return NewConfigError("invalid Revenium API key format", nil)
	}

	Debug("Configuration validation passed")
	return nil
}

// isValidAPIKeyFormat checks if the API key has a valid format
func isValidAPIKeyFormat(key string) bool {
	// Revenium API keys should start with "hak_"
	if len(key) < 4 {
		return false
	}
	return key[:4] == "hak_"
}

// NormalizeReveniumBaseURL normalizes the base URL to a consistent format
// It handles various input formats and returns a normalized base URL without trailing slash
// The endpoint path (/meter/v2/ai/completions) is appended by sendMeteringRequest
func NormalizeReveniumBaseURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}

	if len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}

	if len(baseURL) >= 9 && baseURL[len(baseURL)-9:] == "/meter/v2" {
		return baseURL[:len(baseURL)-9]
	}

	if len(baseURL) >= 6 && baseURL[len(baseURL)-6:] == "/meter" {
		return baseURL[:len(baseURL)-6]
	}

	if len(baseURL) >= 3 && baseURL[len(baseURL)-3:] == "/v2" {
		return baseURL[:len(baseURL)-3]
	}

	// Return the base URL as-is (should be just the domain)
	return baseURL
}
