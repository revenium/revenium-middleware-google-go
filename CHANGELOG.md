# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.0.4] - 2026-01-21

### Added
- **Prompt Capture** - Opt-in feature to capture system prompts, input messages, and output responses for analytics (BACK-396)
- `WithCapturePrompts()` option to enable prompt capture
- `REVENIUM_CAPTURE_PROMPTS` environment variable support
- UTF-8 safe truncation for multi-byte character preservation
- Individual message truncation (before JSON serialization) to prevent invalid JSON

### Fixed
- Boolean parsing consistency for `REVENIUM_CAPTURE_PROMPTS` and `REVENIUM_DEBUG` (accepts both "true" and "1")

## [0.0.3] - 2026-01-17

### Added
- AGENTS.md for AI agent context
- Dynamic version detection using `runtime/debug.ReadBuildInfo()`

## [0.0.1] - 2025-12-16

### Added

- Initial release of Revenium Google Go Middleware
- Support for Google AI (Gemini API) provider
- Support for Vertex AI provider
- Automatic provider detection based on configuration
- Initialize()/GetClient() pattern for client management
- Automatic .env file loading via godotenv
- Context-based metadata management with WithUsageMetadata()
- Comprehensive usage tracking and metering:
  - Token counts (input, output, total, cached, reasoning/thinking tokens)
  - Request timing (duration, time to first token for streaming)
  - Model information and provider detection
  - Stop reason mapping from Google's FinishReason
  - Temperature extraction from GenerateContentConfig
  - Error tracking with detailed error reasons
- Streaming support with automatic metrics tracking:
  - Chunk count tracking
  - Streaming duration measurement
  - Time to first token calculation
  - Accumulated response tracking
- Fire-and-forget metering (non-blocking background processing)
- Debug logging support via REVENIUM_DEBUG environment variable
- Complete metadata support:
  - Organization tracking (organizationId, productId, subscriptionId)
  - Task classification (taskType, agent)
  - Distributed tracing (traceId)
  - Quality metrics (responseQualityScore)
  - Subscriber information (complete object with id, email, credentials)
- Comprehensive examples:
  - Google AI examples: getting-started, basic, streaming, chat, metadata
  - Vertex AI examples: getting-started, basic, streaming, metadata
- Full documentation:
  - README.md with quick start and comprehensive guides
  - DEVELOPMENT.md for contributors and developers
  - CONTRIBUTING.md with contribution guidelines
  - CODE_OF_CONDUCT.md for community standards
  - SECURITY.md for security policy
  - Examples README with detailed example descriptions
- Makefile with convenient commands:
  - make install, test, lint, fmt, clean
  - make run-genai-\* for Google AI examples
  - make run-vertex-\* for Vertex AI examples
  - make build-examples for building all examples
- Stop reason mapping:
  - STOP → END
  - MAX_TOKENS → LENGTH
  - SAFETY → CONTENT_FILTER
  - RECITATION → CONTENT_FILTER
  - OTHER → OTHER
  - FINISH_REASON_UNSPECIFIED → OTHER
- Proper error handling and recovery
- PII protection in debug logs (masked email addresses)
- Automatic transaction ID generation for request tracking

### Technical Details

- Built with Go 1.23+
- Uses google.golang.org/genai v1.26.0 SDK
- Clean architecture with separated concerns:
  - config.go - Configuration management
  - context.go - Context utilities
  - middleware.go - Core middleware logic
  - provider.go - Provider detection
  - stop_reason_mapper.go - Stop reason mapping
  - logger.go - Debug logging
  - errors.go - Error definitions
- Middleware source automatically set to "go"
- Cost type set to "AI"
- Operation type set to "CHAT"
- Provider values: "Google AI" or "Vertex AI"
- ISO 8601 timestamp formatting for all time fields
- Comprehensive test coverage for stop reason mapping

### API Compliance

- Follows Revenium Metering API v2 specification
- Proper field naming and data types
- Correct stop reason enumeration
- Optional fields properly handled (undefined when not available)
- Token counts accurately reported from Google's UsageMetadata
- Streaming metrics properly calculated and reported

[0.0.1]: https://github.com/revenium/revenium-middleware-google-go/releases/tag/v0.0.1
