package revenium

import (
	"context"
)

// contextKey is a type for context keys to avoid collisions
type contextKey string

const (
	usageMetadataKey contextKey = "revenium_usage_metadata"
)

// WithUsageMetadata returns a new context with usage metadata
func WithUsageMetadata(ctx context.Context, metadata map[string]interface{}) context.Context {
	return context.WithValue(ctx, usageMetadataKey, metadata)
}

// GetUsageMetadata retrieves usage metadata from context
func GetUsageMetadata(ctx context.Context) map[string]interface{} {
	if metadata, ok := ctx.Value(usageMetadataKey).(map[string]interface{}); ok {
		return metadata
	}
	return make(map[string]interface{})
}
