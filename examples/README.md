# Revenium Google Go Middleware - Examples

This directory contains examples demonstrating how to use the Revenium Google Go middleware with both Google AI (Gemini API) and Vertex AI.

## Prerequisites

Before running the examples, make sure you have:

1. **Go 1.23+** installed
2. **Revenium API Key** - Get one from [Revenium Dashboard](https://app.revenium.ai)
3. **Google AI API Key** (for Google AI examples) - Get one from [Google AI Studio](https://aistudio.google.com/app/apikey)
4. **Google Cloud Project** (for Vertex AI examples) - Set up at [Google Cloud Console](https://console.cloud.google.com)

## Setup

1. **Clone the repository** (if you haven't already):

   ```bash
   git clone https://github.com/revenium/revenium-middleware-google-go.git
   cd revenium-middleware-google-go
   ```

2. **Install dependencies**:

   ```bash
   go mod download
   ```

3. **Configure environment variables**:

   Create a `.env` file in the project root with your API keys:

   ```bash
   # .env

   # Google AI (Gemini API) Configuration
   GOOGLE_API_KEY=your-google-api-key

   # Vertex AI Configuration (alternative to Google AI)
   GOOGLE_CLOUD_PROJECT=your-project-id
   GOOGLE_CLOUD_LOCATION=us-central1
   GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account-key.json

   # Disable Vertex AI support (set to 1 to disable, 0 to enable)
   REVENIUM_VERTEX_DISABLE=1

   # Revenium Configuration
   REVENIUM_METERING_API_KEY=hak_your_api_key_here
   REVENIUM_METERING_BASE_URL=https://api.revenium.ai  # Optional, defaults to https://api.revenium.ai

   # Debug Configuration (Optional)
   REVENIUM_DEBUG=false  # Set to true to enable debug logging
   ```

   **Note:** The middleware automatically loads `.env` files via `Initialize()`, so no additional configuration is needed.

## Examples

### Google AI (Gemini API) Examples

#### 1. Getting Started

**File:** `google-genai/getting-started/main.go`

The simplest example to get you started with Revenium tracking:

- Initialize the middleware
- Create a basic content generation request
- Display response and usage metrics

**Run:**

```bash
make run-genai-getting-started
# or
go run examples/google-genai/getting-started/main.go
```

**What it does:**

- Loads configuration from environment variables
- Creates a simple content generation request
- Automatically sends metering data to Revenium API
- Displays the response and token usage

---

#### 2. Basic Usage

**File:** `google-genai/basic/main.go`

Demonstrates standard Google AI usage:

- Content generation with metadata
- Simple metadata tracking

**Run:**

```bash
make run-genai-basic
# or
go run examples/google-genai/basic/main.go
```

**What it does:**

- Creates content generation with metadata tracking
- Demonstrates basic metadata usage
- Shows token counting and usage metrics

---

#### 3. Streaming

**File:** `google-genai/streaming/main.go`

Demonstrates streaming responses:

- Real-time token streaming
- Accumulating responses
- Streaming metrics

**Run:**

```bash
make run-genai-streaming
# or
go run examples/google-genai/streaming/main.go
```

**What it does:**

- Creates a streaming content generation request
- Displays tokens as they arrive in real-time
- Tracks streaming metrics including time to first token
- Sends metering data after stream completes

---

#### 4. Chat

**File:** `google-genai/chat/main.go`

Demonstrates multi-turn conversations:

- Chat sessions with history
- Context management
- Conversation tracking

**Run:**

```bash
make run-genai-chat
# or
go run examples/google-genai/chat/main.go
```

**What it does:**

- Demonstrates multi-turn conversation with chat history
- Shows how to maintain conversation context
- Tracks each turn with separate metering

---

#### 5. Metadata

**File:** `google-genai/metadata/main.go`

Demonstrates all available metadata fields:

- Complete metadata structure
- All optional fields documented
- Subscriber information

**Run:**

```bash
make run-genai-metadata
# or
go run examples/google-genai/metadata/main.go
```

**What it does:**

- Shows all available metadata fields
- Demonstrates subscriber tracking
- Includes organization and product tracking

**Metadata fields supported:**

- `traceId` - Session or conversation tracking identifier
- `taskType` - Type of AI task being performed
- `agent` - AI agent or bot identifier
- `organizationId` - Organization identifier
- `productId` - Product or service identifier
- `subscriptionId` - Subscription tier identifier
- `responseQualityScore` - Quality rating (0.0-1.0)
- `subscriber` - Nested subscriber object with `id`, `email`, `credential` (with `name` and `value`)

---

### Vertex AI Examples

#### 1. Getting Started

**File:** `google-vertex/getting-started/main.go`

The simplest Vertex AI example:

- Initialize the middleware with Vertex AI
- Create a basic content generation request
- Display response and usage metrics

**Run:**

```bash
make run-vertex-getting-started
# or
go run examples/google-vertex/getting-started/main.go
```

**What it does:**

- Loads Vertex AI configuration from environment variables
- Creates a simple content generation request using Vertex AI
- Automatically sends metering data to Revenium API
- Displays the response and token usage

---

#### 2. Basic Usage

**File:** `google-vertex/basic/main.go`

Demonstrates standard Vertex AI usage:

- Content generation with metadata
- Simple metadata tracking

**Run:**

```bash
make run-vertex-basic
# or
go run examples/google-vertex/basic/main.go
```

**What it does:**

- Creates content generation with metadata tracking
- Demonstrates basic metadata usage with Vertex AI
- Shows token counting and usage metrics

---

#### 3. Streaming

**File:** `google-vertex/streaming/main.go`

Demonstrates Vertex AI streaming responses:

- Real-time token streaming
- Accumulating responses
- Streaming metrics

**Run:**

```bash
make run-vertex-streaming
# or
go run examples/google-vertex/streaming/main.go
```

**What it does:**

- Creates a streaming content generation request with Vertex AI
- Displays tokens as they arrive in real-time
- Tracks streaming metrics including time to first token
- Sends metering data after stream completes

---

#### 4. Metadata

**File:** `google-vertex/metadata/main.go`

Demonstrates all available metadata fields with Vertex AI:

- Complete metadata structure
- All optional fields documented
- Subscriber information

**Run:**

```bash
make run-vertex-metadata
# or
go run examples/google-vertex/metadata/main.go
```

**What it does:**

- Shows all available metadata fields with Vertex AI
- Demonstrates subscriber tracking
- Includes organization and product tracking

---

## Common Issues

### "Client not initialized" error

**Solution:** Make sure to call `Initialize()` before using `GetClient()`.

### "REVENIUM_METERING_API_KEY is required" error

**Solution:** Set the `REVENIUM_METERING_API_KEY` environment variable in your `.env` file.

### "GOOGLE_API_KEY is required" error (Google AI)

**Solution:** Set the `GOOGLE_API_KEY` environment variable in your `.env` file for Google AI examples.

### "GOOGLE_CLOUD_PROJECT is required" error (Vertex AI)

**Solution:** Set the `GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION`, and `GOOGLE_APPLICATION_CREDENTIALS` environment variables in your `.env` file for Vertex AI examples.

### Environment variables not loading

**Solution:** Make sure your `.env` file is in the project root directory and contains the required variables.

### Google API errors

**Solution:**

- For Google AI: Make sure you have set `GOOGLE_API_KEY` in your `.env` file
- For Vertex AI: Make sure you have set `GOOGLE_CLOUD_PROJECT`, `GOOGLE_CLOUD_LOCATION`, and `GOOGLE_APPLICATION_CREDENTIALS` in your `.env` file

### Debug Mode

Enable detailed logging to troubleshoot issues:

```bash
# In .env file
REVENIUM_DEBUG=true

# Then run examples
make run-genai-getting-started
```

## Next Steps

- Check the [main README](../README.md) for detailed documentation
- Visit the [Revenium Dashboard](https://app.revenium.ai) to view your metering data
- See [.env.example](../.env.example) for all configuration options

## Support

For issues or questions:

- Documentation: https://docs.revenium.io
- Email: support@revenium.io
