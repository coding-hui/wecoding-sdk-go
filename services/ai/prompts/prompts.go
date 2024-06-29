// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package prompts

import "github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

// Formatter is an interface for formatting a map of values into a string.
type Formatter interface {
	Format(values map[string]any) (string, error)
}

// MessageFormatter is an interface for formatting a map of values into a list
// of messages.
type MessageFormatter interface {
	FormatMessages(values map[string]any) ([]llms.ChatMessage, error)
	GetInputVariables() []string
}

// FormatPrompter is an interface for formatting a map of values into a prompt.
type FormatPrompter interface {
	FormatPrompt(values map[string]any) (llms.PromptValue, error)
	GetInputVariables() []string
}
