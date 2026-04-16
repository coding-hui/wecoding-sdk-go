# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the WeCoding Go Management SDK - a Go SDK for IAM (Identity and Access Management) and AI/LLM services. It provides both REST API clients for IAM operations and abstractions for working with LLMs.

## Build and Development Commands

```bash
# Run all tests
make test

# Run a single test
go test ./services/ai/llms/...

# Format code (gofmt, goimports, golines with max-len=120)
make format

# Lint code
make lint

# Verify copyright headers
make verify-copyright

# Add copyright headers if missing
make add-copyright
```

## Architecture

### Client Hierarchy (Kubernetes-style pattern)

```
Clientset → IamClient → APIV1Client/AuthzV1Client → RESTClient → gorequest
```

- `services/clientset.go` - Top-level clientset containing all service clients
- `services/iam/iam_client.go` - IAM client grouping APIV1 and AuthzV1
- `services/iam/apiserver/v1/` - API server operations (authentication, users)
- `services/iam/authz/v1/` - Authorization operations
- `rest/` - HTTP REST client implementation using gorequest

### AI Services Structure

```
services/ai/
├── callbacks/     # LLM callbacks (simple, log, log_stream, combining)
├── llms/          # LLM implementations (chat messages, content generation)
│   ├── openai/    # OpenAI-compatible client
│   └── volcengine/ # Volcengine client
├── outputparser/  # Response parsing (boolean, structured, regex, etc.)
└── prompts/       # Prompt templates and formatting
```

### REST Layer

- `rest/client.go` - RESTClient with authentication support (Basic, Bearer, SecretID/SecretKey)
- `rest/config.go` - Configuration for REST clients
- `rest/request.go` - Request building and execution

## Code Conventions

- **Copyright header**: All files must have the MIT license header (see `boilerplate.txt`)
- **Line length**: `lll` linter allows lines up to 240 characters
- **Max function complexity**: 30 (gocyclo, gocognit)
- **Max function length**: 150 lines (funlen)
- **Import prefix**: `github.com/coding-hui/` for internal imports
- **Test files**: Named `*_test.go`, tests use `testify` assert package

## Key Dependencies

- `github.com/coding-hui/common` - Common utilities
- `github.com/coding-hui/iam` - IAM types
- `github.com/sashabaranov/go-openai` - OpenAI client
- `github.com/volcengine/volcengine-go-sdk` - Volcengine SDK
