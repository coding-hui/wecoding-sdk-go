// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package llms

import (
	"context"

	"github.com/coding-hui/common/errors"
)

// Model is an interface multi-modal models implement.
type Model interface {
	// GenerateContent asks the model to generate content from a sequence of
	// messages. It's the most general interface for multi-modal LLMs that support
	// chat-like interactions.
	GenerateContent(ctx context.Context, messages []MessageContent, options ...CallOption) (*ContentResponse, error)
}

// GenerateFromSinglePrompt is a convenience function for calling an LLM with
// a single string prompt, expecting a single string response. It's useful for
// simple, string-only interactions and provides a slightly more ergonomic API
// than the more general [llms.Model.GenerateContent].
func GenerateFromSinglePrompt(ctx context.Context, llm Model, prompt string, options ...CallOption) (string, error) {
	msg := MessageContent{
		Role:  ChatMessageTypeHuman,
		Parts: []ContentPart{TextContent{prompt}},
	}

	resp, err := llm.GenerateContent(ctx, []MessageContent{msg}, options...)
	if err != nil {
		return "", err
	}

	choices := resp.Choices
	if len(choices) < 1 {
		return "", errors.New("empty response from model")
	}
	c1 := choices[0]
	return c1.Content, nil
}
