# AGENTS.md

> Machine-readable instructions for AI agents. Human docs: [README.md](README.md)

## Project Context

**Type**: Go middleware for Google AI (Gemini) API
**Purpose**: Add Revenium metering to Google Gemini API calls
**Stack**: Go 1.21+
**Module**: `github.com/revenium/revenium-middleware-google-go`

## Commands

```bash
# Build
go build ./...

# Test
go test ./...

# Run example
cd examples && go run getting_started.go
```

## Architecture

```
revenium/
├── client.go      # Gemini client wrapper
├── config.go      # Configuration and validation
├── context.go     # Context metadata handling
├── errors.go      # Error types
├── logger.go      # Logging utilities
├── metering.go    # Revenium metering (fire-and-forget)
└── middleware.go  # Core middleware logic
```

## Critical Constraints

1. **Dynamic versioning** - Use `GetMiddlewareSource()` for metering payloads (never hardcode)
2. **Billing fields at TOP LEVEL** - `inputTokens`, `outputTokens` NOT in attributes
3. **Fire-and-forget metering** - Use goroutines, never block main request
4. **Auth header** - Use `x-api-key` (not `Authorization: Bearer`)

## Environment Variables

```bash
GOOGLE_API_KEY=...                 # Required: Google AI API key
REVENIUM_METERING_API_KEY=hak_...  # Required: Revenium key (starts with hak_)
REVENIUM_DEBUG=true                # Optional: Enable debug logging
```

## Metering Endpoints

| Type | Endpoint | Key Fields |
|------|----------|------------|
| Chat | `/meter/v1/ai/meter` | `inputTokens`, `outputTokens` |

## Common Errors

| Error | Fix |
|-------|-----|
| `package not found` | `go mod tidy` |
| `metering not tracking` | Check `REVENIUM_METERING_API_KEY` |
| `Google auth failed` | Check `GOOGLE_API_KEY` |

## References

- [Revenium Docs](https://docs.revenium.io)
- [Google AI Docs](https://ai.google.dev/docs)
- [AGENTS.md Spec](https://agents.md/)
