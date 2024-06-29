// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package openai

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/callbacks"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

func newTestClient(t *testing.T, opts ...Option) llms.Model {
	t.Helper()
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey == "" {
		t.Skip("OPENAI_API_KEY not set")
		return nil
	}

	llm, err := New(opts...)
	llm.CallbacksHandler = callbacks.LogHandler{}

	require.NoError(t, err)
	return llm
}

func TestMultiContentText(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, WithAPIType(APITypeAzure))

	parts := []llms.ContentPart{
		llms.TextPart("你是谁"),
		llms.TextPart("你可以干什么"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
}

func TestMultiContentTextChatSequence(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, WithBaseURL("https://api.closeai-proxy.xyz/v1"), WithModel("gpt-3.5-turbo"))

	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Name some countries")},
		},
		{
			Role:  llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{llms.TextPart("Spain and Lesotho")},
		},
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{llms.TextPart("Which if these is larger?")},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content)
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "spain.*larger", strings.ToLower(c1.Content))
}

func TestMultiContentImage(t *testing.T) {
	t.Parallel()

	llm := newTestClient(t, WithBaseURL("https://api.closeai-proxy.xyz/v1"), WithModel("gpt-4o"))

	parts := []llms.ContentPart{
		llms.ImageURLPart("https://github.com/tmc/langchaingo/blob/main/docs/static/img/parrot-icon.png?raw=true"),
		llms.TextPart("describe this image in detail"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithMaxTokens(300))
	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "parrot", strings.ToLower(c1.Content))
}

func TestWithStreaming(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, WithBaseURL("https://api.closeai-proxy.xyz/v1"), WithModel("gpt-3.5-turbo"))

	parts := []llms.ContentPart{
		llms.TextPart("I'm a pomeranian"),
		llms.TextPart("Tell me more about my taxonomy"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	var sb strings.Builder
	rsp, err := llm.GenerateContent(context.Background(), content,
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			sb.Write(chunk)
			return nil
		}),
	)

	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Regexp(t, "dog|canid", strings.ToLower(c1.Content))
	assert.Regexp(t, "dog|canid", strings.ToLower(sb.String()))
}

//nolint:lll
func TestToolCall(t *testing.T) {
	t.Parallel()
	llm := newTestClient(t, WithBaseURL("https://api.closeai-proxy.xyz/v1"), WithModel("gpt-3.5-turbo"))

	parts := []llms.ContentPart{
		llms.TextPart("What is the weather like in Boston?"),
	}
	content := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeHuman,
			Parts: parts,
		},
	}

	tools := []llms.Tool{
		{
			"function",
			&llms.FunctionDefinition{
				Name:        "getCurrentWeather",
				Description: "Get the current weather in a given location",
				Parameters:  json.RawMessage(`{"type": "object", "properties": {"location": {"type": "string", "description": "The city and state, e.g. San Francisco, CA"}, "unit": {"type": "string", "enum": ["celsius", "fahrenheit"]}}, "required": ["location"]}`),
			},
		},
	}

	rsp, err := llm.GenerateContent(context.Background(), content, llms.WithTools(tools))

	require.NoError(t, err)

	assert.NotEmpty(t, rsp.Choices)
	c1 := rsp.Choices[0]
	assert.Equal(t, "tool_calls", c1.StopReason)
	assert.NotNil(t, c1.FuncCall)
}
