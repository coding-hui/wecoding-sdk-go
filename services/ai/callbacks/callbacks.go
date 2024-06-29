// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package callbacks

import (
	"context"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

// Handler is the interface that allows for hooking into specific parts of an
// LLM application.
//
//nolint:all
type Handler interface {
	HandleText(ctx context.Context, text string)
	HandleLLMStart(ctx context.Context, prompts []string)
	HandleLLMGenerateContentStart(ctx context.Context, ms []llms.MessageContent)
	HandleLLMGenerateContentEnd(ctx context.Context, res *llms.ContentResponse)
	HandleLLMError(ctx context.Context, err error)
	HandleChainStart(ctx context.Context, inputs map[string]any)
	HandleChainEnd(ctx context.Context, outputs map[string]any)
	HandleChainError(ctx context.Context, err error)
	HandleToolStart(ctx context.Context, input string)
	HandleToolEnd(ctx context.Context, output string)
	HandleToolError(ctx context.Context, err error)
	HandleStreamingFunc(ctx context.Context, chunk []byte)
}

// HandlerHaver is an interface used to get callbacks handler.
type HandlerHaver interface {
	GetCallbackHandler() Handler
}
