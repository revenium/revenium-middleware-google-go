# Revenium Middleware for Google (Go)

A lightweight, production-ready middleware that adds **Revenium metering and tracking** to Google AI (Gemini API) and Vertex AI API calls.

[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue)](https://golang.org/)
[![Documentation](https://img.shields.io/badge/docs-revenium.io-blue)](https://docs.revenium.io)
[![Website](https://img.shields.io/badge/website-revenium.ai-blue)](https://www.revenium.ai)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Dual Provider Support** - Works with both Google AI (Gemini API) and Vertex AI
- **Automatic Metering** - Tracks all API calls with detailed usage metrics
- **Streaming Support** - Full support for streaming responses
- **Custom Metadata** - Add custom tracking metadata to any request
- **Production Ready** - Battle-tested and optimized for production use
- **Type Safe** - Built with Go's strong typing system

## Getting Started (5 minutes)

### Step 1: Create Your Project

```bash
mkdir my-google-ai-project
cd my-google-ai-project
go mod init my-google-ai-project
```

### Step 2: Install Dependencies

```bash
go get google.golang.org/genai
go get github.com/revenium/revenium-middleware-google-go
go mod tidy
```

This installs both the Google AI SDK and the Revenium middleware.

### Step 3: Create Environment File

Create a `.env` file in your project root with your API keys:

```bash
# .env

# Google AI (Gemini API) Configuration
GOOGLE_API_KEY=your-api-key-here

# Vertex AI Configuration (alternative to Google AI)
GOOGLE_CLOUD_PROJECT=your-project-id-here
GOOGLE_CLOUD_LOCATION=your-location-here
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json

# Disable Vertex AI support (set to 1 to disable, 0 to enable)
REVENIUM_VERTEX_DISABLE=1

# Revenium Configuration
REVENIUM_METERING_API_KEY=hak_your_api_key_here
REVENIUM_METERING_BASE_URL=https://api.revenium.ai

# Debug Configuration (Optional)
REVENIUM_DEBUG=false
```

**Replace the placeholder values with your actual keys!**

For a complete list of all available environment variables, see the [Configuration Options](#configuration-options) section below.

> **Note**: `REVENIUM_METERING_BASE_URL` defaults to `https://api.revenium.ai` and doesn't need to be set unless using a different environment.

## Examples

This repository includes runnable examples demonstrating how to use the Revenium middleware with Google AI (Gemini API) and Vertex AI:

- [Google AI (Gemini API) Examples](https://github.com/revenium/revenium-middleware-google-go/tree/HEAD/examples/google-genai)
- [Vertex AI Examples](https://github.com/revenium/revenium-middleware-google-go/tree/HEAD/examples/google-vertex)

### Run examples after setup:

```bash
# Google AI (Gemini API) Examples
make run-genai-getting-started
make run-genai-basic
make run-genai-streaming
make run-genai-chat
make run-genai-metadata

# Vertex AI Examples
make run-vertex-getting-started
make run-vertex-basic
make run-vertex-streaming
make run-vertex-metadata
```

## What Gets Tracked

The middleware automatically captures comprehensive usage data:

### **Usage Metrics (Automatic)**

- **Token Counts** - Input tokens, output tokens, total tokens, reasoning tokens, cached tokens
- **Model Information** - Model name, provider (Google AI or Vertex AI)
- **Request Timing** - Request duration, response time, time to first token (streaming)
- **Streaming Metrics** - Chunk count, streaming duration
- **Stop Reason** - Automatically mapped from Google's `FinishReason` to Revenium's standardized stop reasons
- **Temperature** - Automatically extracted from `GenerateContentConfig`
- **Error Tracking** - Failed requests with error reasons

### **Business Context (Optional via Metadata)**

- **Organization Data** - Organization ID, product ID, subscription ID
- **Task Classification** - Task type, agent identifier
- **Tracing** - Trace ID for distributed tracing
- **Quality Metrics** - Response quality score
- **Subscriber Information** - Complete subscriber object with ID, email, and credentials

### **Technical Details**

- **API Endpoints** - Content generation (streaming and non-streaming)
- **Request Types** - Streaming vs non-streaming
- **Provider Detection** - Automatic detection of Google AI vs Vertex AI
- **Middleware Source** - Automatically set to "go"
- **Transaction ID** - Unique ID for each request

## Environment Variables

### Required

```bash
REVENIUM_METERING_API_KEY=hak_your_api_key_here
GOOGLE_API_KEY=your-api-key-here  # For Google AI (Gemini API)
REVENIUM_VERTEX_DISABLE=1  # Set to 1 to disable Vertex AI support
```

### Optional

```bash
REVENIUM_DEBUG=false  # Set to true to enable debug logging
REVENIUM_METERING_BASE_URL=https://api.revenium.ai  # Optional, defaults to https://api.revenium.ai
GOOGLE_CLOUD_PROJECT=your-project-id-here  # For Vertex AI
GOOGLE_CLOUD_LOCATION=your-location-here  # For Vertex AI
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json  # For Vertex AI

```

## VERTEX AI Configuration

To use Vertex AI with the middleware:

### 1. Configure VERTEX AI

```bash
# 1. Create a service account with the "Vertex AI User" role
# 2. Download the service account key as a JSON file
# 3. Set the following environment variables:
GOOGLE_CLOUD_PROJECT=your-project-id-here
GOOGLE_CLOUD_LOCATION=your-location-here
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json
```

### 2. Enable Vertex AI

```bash
REVENIUM_VERTEX_DISABLE=0  # Set to 0 to enable Vertex AI support
```

See the getting started example for Vertex AI [here](https://github.com/revenium/revenium-middleware-google-go/tree/HEAD/examples/google-vertex/getting-started)

## API Overview

- **`Initialize()`** - Initialize the middleware from environment variables
- **`GetClient()`** - Get the global Revenium client instance
- **`NewReveniumGoogle(ctx, cfg)`** - Create a new client with explicit configuration
- **`WithUsageMetadata(ctx, metadata)`** - Add custom metadata to a request context
- **`Close()`** - Wait for all pending metering requests to complete

**For complete API documentation and usage examples, see [`examples/README.md`](https://github.com/revenium/revenium-middleware-google-go/tree/HEAD/examples/README.md).**

## Metadata Fields

The middleware supports the following optional metadata fields for tracking:

| Field                   | Type   | Description                                                |
| ----------------------- | ------ | ---------------------------------------------------------- |
| `traceId`               | string | Unique identifier for session or conversation tracking     |
| `taskType`              | string | Type of AI task being performed (e.g., "chat", "analysis") |
| `agent`                 | string | AI agent or bot identifier                                 |
| `organizationId`        | string | Organization or company identifier                         |
| `productId`             | string | Your product or feature identifier                         |
| `subscriptionId`        | string | Subscription plan identifier                               |
| `responseQualityScore`  | number | Custom quality rating (0.0-1.0)                            |
| `subscriber.id`         | string | Unique user identifier                                     |
| `subscriber.email`      | string | User email address                                         |
| `subscriber.credential` | object | Authentication credential (`name` and `value` fields)      |

**All metadata fields are optional.** For complete metadata documentation and usage examples, see:

- [`examples/README.md`](https://github.com/revenium/revenium-middleware-google-go/tree/HEAD/examples/README.md) - All usage examples
- [Revenium API Reference](https://revenium.readme.io/reference/meter_ai_completion) - Complete API documentation

## How It Works

1. **Initialize**: Call `Initialize()` to set up the middleware with your configuration
2. **Get Client**: Call `GetClient()` to get a wrapped Google AI/Vertex AI client instance
3. **Make Requests**: Use the client normally - all requests are automatically tracked
4. **Async Tracking**: Usage data is sent to Revenium in the background (fire-and-forget)
5. **Transparent Response**: Original Google AI/Vertex AI responses are returned unchanged
6. **Graceful Shutdown**: Call `Close()` to wait for all pending metering requests

The middleware never blocks your application - if Revenium tracking fails, your Google AI/Vertex AI requests continue normally.

**Supported APIs:**

- Content Generation API (`client.Models().GenerateContent()`)
- Streaming API (`client.Models().GenerateContentStream()`)
- Both Google AI (Gemini API) and Vertex AI providers

## Troubleshooting

### Common Issues

**No tracking data appears:**

1. Verify environment variables are set correctly in `.env`
2. Enable debug logging by setting `REVENIUM_DEBUG=true` in `.env`
3. Check console for `[Revenium]` log messages
4. Verify your `REVENIUM_METERING_API_KEY` is valid

**Client not initialized error:**

- Make sure you call `Initialize()` before `GetClient()`
- Check that your `.env` file is in the project root
- Verify `REVENIUM_METERING_API_KEY` is set

**Google AI API errors:**

- Verify `GOOGLE_API_KEY` is set correctly
- Ensure you're using a valid model name (e.g., `gemini-2.0-flash-exp`)

**Vertex AI API errors:**

- Verify all three Vertex AI variables are set: `GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION`, `GOOGLE_APPLICATION_CREDENTIALS`
- Check that the service account JSON file exists at the specified path
- Ensure the service account has the necessary permissions

**Wrong provider detected:**

- If you have both Google AI and Vertex AI credentials configured, the middleware will auto-detect based on `GOOGLE_CLOUD_PROJECT`
- To force Google AI, set `REVENIUM_VERTEX_DISABLE=1`
- To force Vertex AI, set `REVENIUM_VERTEX_DISABLE=0` (or leave unset)

### Debug Mode

Enable detailed logging by adding to your `.env`:

```env
REVENIUM_DEBUG=true
```

### Getting Help

If issues persist:

1. Enable debug logging (`REVENIUM_DEBUG=true`)
2. Check the [`examples/`](https://github.com/revenium/revenium-middleware-google-go/tree/HEAD/examples) directory for working examples
3. Review [`examples/README.md`](https://github.com/revenium/revenium-middleware-google-go/tree/HEAD/examples/README.md) for detailed setup instructions
4. Contact support@revenium.io with debug logs

## Supported Models

This middleware works with any Google AI or Vertex AI model. For the complete model list, see:

- [Google AI Models Documentation](https://ai.google.dev/gemini-api/docs/models)
- [Vertex AI Models Documentation](https://cloud.google.com/vertex-ai/generative-ai/docs/learn/models)

### API Support Matrix

The following table shows what has been tested and verified with working examples:

| Feature               | Google AI | Vertex AI |
| --------------------- | --------- | --------- |
| **Basic Usage**       | Yes       | Yes       |
| **Streaming**         | Yes       | Yes       |
| **Multi-turn Chat**   | Yes       | -         |
| **Metadata Tracking** | Yes       | Yes       |
| **Token Counting**    | Yes       | Yes       |

**Note:** "Yes" = Tested with working examples in [`examples/`](https://github.com/revenium/revenium-middleware-google-go/tree/HEAD/examples) directory

## Requirements

- Go 1.23+
- Revenium API key
- Google AI API key (for Google AI) OR Google Cloud project with Vertex AI enabled (for Vertex AI)

## Documentation

For detailed documentation, visit [docs.revenium.io](https://docs.revenium.io)

## License

This project is licensed under the MIT License - see the [LICENSE](https://github.com/revenium/revenium-middleware-google-go/blob/HEAD/LICENSE) file for details.

## Support

For issues, feature requests, or contributions:

- **GitHub Repository**: [revenium/revenium-middleware-google-go](https://github.com/revenium/revenium-middleware-google-go)
- **Issues**: [Report bugs or request features](https://github.com/revenium/revenium-middleware-google-go/issues)
- **Documentation**: [docs.revenium.io](https://docs.revenium.io)
- **Contact**: Reach out to the Revenium team for additional support

---

**Built by Revenium**
