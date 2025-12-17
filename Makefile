.PHONY: help install test lint fmt clean build-examples
.PHONY: run-genai-getting-started run-genai-basic run-genai-streaming run-genai-chat run-genai-metadata
.PHONY: run-vertex-getting-started run-vertex-basic run-vertex-streaming run-vertex-metadata

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install: ## Install dependencies
	go mod download
	go mod tidy

test: ## Run tests
	go test -v ./...

lint: ## Run linter
	go vet ./...
	go fmt ./...

fmt: ## Format code
	go fmt ./...

clean: ## Clean build artifacts
	go clean
	rm -rf bin/

# Google GenAI Examples
run-genai-getting-started: ## Run Google GenAI getting started example
	go run examples/google-genai/getting-started/main.go

run-genai-basic: ## Run Google GenAI basic example
	go run examples/google-genai/basic/main.go

run-genai-streaming: ## Run Google GenAI streaming example
	go run examples/google-genai/streaming/main.go

run-genai-chat: ## Run Google GenAI chat example
	go run examples/google-genai/chat/main.go

run-genai-metadata: ## Run Google GenAI metadata example
	go run examples/google-genai/metadata/main.go

# Vertex AI Examples
run-vertex-getting-started: ## Run Vertex AI getting started example
	go run examples/google-vertex/getting-started/main.go

run-vertex-basic: ## Run Vertex AI basic example
	go run examples/google-vertex/basic/main.go

run-vertex-streaming: ## Run Vertex AI streaming example
	go run examples/google-vertex/streaming/main.go

run-vertex-metadata: ## Run Vertex AI metadata example
	go run examples/google-vertex/metadata/main.go

build-examples: ## Build all examples
	@mkdir -p bin/google-genai bin/google-vertex
	go build -o bin/google-genai/getting-started examples/google-genai/getting-started/main.go
	go build -o bin/google-genai/basic examples/google-genai/basic/main.go
	go build -o bin/google-genai/streaming examples/google-genai/streaming/main.go
	go build -o bin/google-genai/chat examples/google-genai/chat/main.go
	go build -o bin/google-genai/metadata examples/google-genai/metadata/main.go
	go build -o bin/google-vertex/getting-started examples/google-vertex/getting-started/main.go
	go build -o bin/google-vertex/basic examples/google-vertex/basic/main.go
	go build -o bin/google-vertex/streaming examples/google-vertex/streaming/main.go
	go build -o bin/google-vertex/metadata examples/google-vertex/metadata/main.go
