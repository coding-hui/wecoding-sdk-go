// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	sdk "github.com/sashabaranov/go-openai"

	"github.com/coding-hui/common/errors"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/callbacks"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

var (
	ErrEmptyResponse              = errors.New("no response")
	ErrMissingToken               = errors.New("missing the OpenAI API key, set it in the OPENAI_API_KEY environment variable") //nolint:lll
	ErrMissingAzureModel          = errors.New("model needs to be provided when using Azure API")
	ErrMissingAzureEmbeddingModel = errors.New("embeddings model needs to be provided when using Azure API")

	ErrUnexpectedResponseLength = errors.New("unexpected length of response")
)

type Model struct {
	client        *sdk.Client
	clientOptions *clientOptions

	CallbacksHandler callbacks.Handler
}

const (
	RoleSystem    = "system"
	RoleAssistant = "assistant"
	RoleUser      = "user"
	RoleFunction  = "function"
	RoleTool      = "tool"
)

var _ llms.Model = (*Model)(nil)

func New(opts ...Option) (*Model, error) {
	options := &clientOptions{
		token:        os.Getenv(tokenEnvVarName),
		model:        os.Getenv(modelEnvVarName),
		organization: os.Getenv(organizationEnvVarName),
		baseURL:      getEnvs(baseURLEnvVarName, baseAPIBaseEnvVarName),
		apiType:      APITypeOpenAI,
		httpClient:   http.DefaultClient,
	}

	for _, opt := range opts {
		opt(options)
	}

	if len(options.token) == 0 {
		return nil, ErrMissingToken
	}

	clientCfg := sdk.DefaultConfig(options.token)
	// set of options needed for Azure client
	if isAzureApi(options.apiType) && options.apiVersion == "" {
		options.apiVersion = DefaultAPIVersion
		if options.model == "" {
			return nil, ErrMissingAzureModel
		}
		clientCfg = sdk.DefaultAzureConfig(options.token, options.baseURL)
	}

	clientCfg.HTTPClient = options.httpClient
	clientCfg.BaseURL = options.baseURL
	clientCfg.OrgID = options.organization
	clientCfg.APIType = options.apiType
	clientCfg.APIVersion = options.apiVersion

	return &Model{
		clientOptions: options,
		client:        sdk.NewClientWithConfig(clientCfg),
	}, nil
}

// GenerateContent implements the Model interface.
func (o *Model) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, goerr113, funlen
	startTime := time.Now()

	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	if len(opts.Model) == 0 {
		opts.Model = o.clientOptions.model
	}

	chatMsgs := make([]sdk.ChatCompletionMessage, 0, len(messages))
	for _, mc := range messages {
		content, multiContent := messagePartsFromParts(mc.Parts)
		msg := sdk.ChatCompletionMessage{Content: content}
		if opts.SupportMultiContent {
			msg = sdk.ChatCompletionMessage{MultiContent: multiContent}
		}
		switch mc.Role {
		case llms.ChatMessageTypeSystem:
			msg.Role = RoleSystem
		case llms.ChatMessageTypeAI:
			msg.Role = RoleAssistant
		case llms.ChatMessageTypeHuman:
			msg.Role = RoleUser
		case llms.ChatMessageTypeGeneric:
			msg.Role = RoleUser
		case llms.ChatMessageTypeFunction:
			msg.Role = RoleFunction
		case llms.ChatMessageTypeTool:
			msg.Role = RoleTool
			// parse mc.Parts (which should have one entry of type ToolCallResponse) and populate msg.Content and msg.ToolCallID
			if len(mc.Parts) != 1 {
				return nil, fmt.Errorf("expected exactly one part for role %v, got %v", mc.Role, len(mc.Parts))
			}
			switch p := mc.Parts[0].(type) {
			case llms.ToolCallResponse:
				msg.ToolCallID = p.ToolCallID
				msg.Content = p.Content
			default:
				return nil, fmt.Errorf("expected part of type ToolCallResponse for role %v, got %T", mc.Role, mc.Parts[0])
			}

		default:
			return nil, fmt.Errorf("role %v not supported", mc.Role)
		}

		// Here we extract tool calls from the message and populate the ToolCalls field.
		msg.ToolCalls = extractToolParts(mc.Parts)

		chatMsgs = append(chatMsgs, msg)
	}
	req := sdk.ChatCompletionRequest{
		Model:            opts.Model,
		Stop:             opts.StopWords,
		Messages:         chatMsgs,
		Temperature:      float32(opts.Temperature),
		MaxTokens:        opts.MaxTokens,
		N:                opts.N,
		FrequencyPenalty: float32(opts.FrequencyPenalty),
		PresencePenalty:  float32(opts.PresencePenalty),

		ToolChoice: opts.ToolChoice,
		Seed:       &opts.Seed,
	}
	if opts.JSONMode {
		req.ResponseFormat = ChatCompletionResponseFormatJSON
	}

	// if opts.Tools is not empty, append them to req.Tools
	for _, tool := range opts.Tools {
		t, err := toolFromTool(tool)
		if err != nil {
			return nil, fmt.Errorf("failed to convert llms tool to openai tool: %w", err)
		}
		req.Tools = append(req.Tools, t)
	}

	if opts.StreamingFunc != nil {
		return o.parseStreamingChatResponse(ctx, startTime, req, opts.StreamingFunc)
	}

	result, err := o.client.CreateChatCompletion(ctx, req)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}
	if len(result.Choices) == 0 {
		return nil, ErrEmptyResponse
	}

	choices := make([]*llms.ContentChoice, len(result.Choices))
	for i, c := range result.Choices {
		choices[i] = &llms.ContentChoice{
			Content:    c.Message.Content,
			StopReason: fmt.Sprint(c.FinishReason),
		}

		// Legacy function call handling
		if c.FinishReason == "function_call" {
			choices[i].FuncCall = &llms.FunctionCall{
				Name:      c.Message.FunctionCall.Name,
				Arguments: c.Message.FunctionCall.Arguments,
			}
		}
		for _, tool := range c.Message.ToolCalls {
			choices[i].ToolCalls = append(choices[i].ToolCalls, llms.ToolCall{
				ID:   tool.ID,
				Type: string(tool.Type),
				FunctionCall: &llms.FunctionCall{
					Name:      tool.Function.Name,
					Arguments: tool.Function.Arguments,
				},
			})
		}
		// populate legacy single-function call field for backwards compatibility
		if len(choices[i].ToolCalls) > 0 {
			choices[i].FuncCall = choices[i].ToolCalls[0].FunctionCall
		}
	}

	response := &llms.ContentResponse{
		Choices: choices,
		Usage:   getUsage(&result.Usage, startTime, 0),
	}
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}
	return response, nil
}

func (o *Model) parseStreamingChatResponse(
	ctx context.Context,
	startTime time.Time,
	req sdk.ChatCompletionRequest,
	streamingFunc func(ctx context.Context, chunk []byte) error,
) (*llms.ContentResponse, error) {
	req.Stream = true
	req.StreamOptions = &sdk.StreamOptions{
		IncludeUsage: true,
	}

	stream, err := o.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	responseChan := make(chan sdk.ChatCompletionStreamResponse)
	go func() {
		defer close(responseChan)
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				chunk.Choices = append(chunk.Choices, sdk.ChatCompletionStreamChoice{
					Delta: sdk.ChatCompletionStreamChoiceDelta{
						Content: fmt.Errorf("error decoding streaming response: %w", err).Error(),
					},
				})
				responseChan <- chunk
				return
			}
			responseChan <- chunk
		}
	}()

	response, err := o.combineStreamingChatResponse(ctx, startTime, responseChan, streamingFunc)
	if err != nil {
		return nil, err
	}
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}
	return response, nil
}

func (o *Model) combineStreamingChatResponse(
	ctx context.Context,
	startTime time.Time,
	responseChan chan sdk.ChatCompletionStreamResponse,
	streamingFunc func(ctx context.Context, chunk []byte) error,
) (*llms.ContentResponse, error) {
	var hasFirstToken bool
	var firstTokenTime time.Duration
	response := llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{},
		},
		Usage: llms.Usage{},
	}

	for streamResponse := range responseChan {
		if streamResponse.Usage != nil {
			response.Usage = getUsage(streamResponse.Usage, startTime, firstTokenTime)
		}

		if len(streamResponse.Choices) == 0 {
			continue
		}
		choice := streamResponse.Choices[0]
		chunk := []byte(choice.Delta.Content)
		response.Choices[0].Content += choice.Delta.Content
		response.Choices[0].StopReason = fmt.Sprint(choice.FinishReason)

		if !hasFirstToken && choice.Delta.Content != "" {
			firstTokenTime = time.Since(startTime)
			hasFirstToken = true
		}

		if len(choice.Delta.ToolCalls) > 0 {
			chunk, response.Choices[0].ToolCalls = updateToolCalls(response.Choices[0].ToolCalls, choice.Delta.ToolCalls)
		}

		if streamingFunc != nil {
			err := streamingFunc(ctx, chunk)
			if err != nil {
				return nil, fmt.Errorf("streaming func returned an error: %w", err)
			}
		}
	}
	return &response, nil
}

func updateToolCalls(tools []llms.ToolCall, delta []sdk.ToolCall) ([]byte, []llms.ToolCall) {
	if len(delta) == 0 {
		return []byte{}, tools
	}
	for _, t := range delta {
		// if we have arguments append to the last Tool call
		if t.Type == `` && t.Function.Arguments != `` {
			lindex := len(tools) - 1
			if lindex < 0 {
				continue
			}

			tools[lindex].FunctionCall.Arguments += t.Function.Arguments
			continue
		}

		// Otherwise, this is a new tool call, append that to the stack
		tools = append(tools, llms.ToolCall{
			ID:   t.ID,
			Type: fmt.Sprint(t.Type),
			FunctionCall: &llms.FunctionCall{
				Name:      t.Function.Name,
				Arguments: t.Function.Arguments,
			},
		})
	}

	chunk, _ := json.Marshal(delta) // nolint:errchkjson

	return chunk, tools
}

func messagePartsFromParts(parts []llms.ContentPart) (string, []sdk.ChatMessagePart) {
	var content []sdk.ChatMessagePart
	fullContent := ""
	for _, part := range parts {
		switch p := part.(type) {
		case llms.TextContent:
			content = append(content, sdk.ChatMessagePart{
				Type:     sdk.ChatMessagePartTypeText,
				Text:     p.Text,
				ImageURL: nil,
			})
			fullContent += p.Text
		case llms.ImageURLContent:
			content = append(content, sdk.ChatMessagePart{
				Type: sdk.ChatMessagePartTypeText,
				ImageURL: &sdk.ChatMessageImageURL{
					URL:    p.URL,
					Detail: sdk.ImageURLDetail(p.Detail),
				},
			})
		}
	}
	return fullContent, content
}

// toolFromTool converts an llms.Tool to a Tool.
func toolFromTool(t llms.Tool) (sdk.Tool, error) {
	tool := sdk.Tool{
		Type: sdk.ToolType(t.Type),
	}
	switch t.Type {
	case string(sdk.ToolTypeFunction):
		tool.Function = &sdk.FunctionDefinition{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			Parameters:  t.Function.Parameters,
		}
	default:
		return sdk.Tool{}, fmt.Errorf("tool type %v not supported", t.Type)
	}
	return tool, nil
}

// toolCallsFromToolCalls converts a slice of llms.ToolCall to a slice of ToolCall.
func toolCallsFromToolCalls(tcs []llms.ToolCall) []sdk.ToolCall {
	toolCalls := make([]sdk.ToolCall, len(tcs))
	for i, tc := range tcs {
		toolCalls[i] = toolCallFromToolCall(tc)
	}
	return toolCalls
}

// toolCallFromToolCall converts an llms.ToolCall to a ToolCall.
func toolCallFromToolCall(tc llms.ToolCall) sdk.ToolCall {
	return sdk.ToolCall{
		ID:   tc.ID,
		Type: sdk.ToolType(tc.Type),
		Function: sdk.FunctionCall{
			Name:      tc.FunctionCall.Name,
			Arguments: tc.FunctionCall.Arguments,
		},
	}
}

// extractToolParts extracts the tool parts from a message.
func extractToolParts(parts []llms.ContentPart) []sdk.ToolCall {
	var toolCalls []sdk.ToolCall
	for _, part := range parts {
		switch p := part.(type) {
		case llms.ToolCall:
			toolCalls = append(toolCalls, toolCallFromToolCall(p))
		}
	}
	return toolCalls
}

func getEnvs(keys ...string) string {
	for _, key := range keys {
		val, ok := os.LookupEnv(key)
		if ok {
			return val
		}
	}
	return ""
}

func isAzureApi(apiType APIType) bool {
	return apiType == APITypeAzure || apiType == APITypeAzureAD
}

func getUsage(res *sdk.Usage, startTime time.Time, firstTokenTime time.Duration) llms.Usage {
	if res == nil {
		return llms.Usage{}
	}

	totalDuration := time.Since(startTime)
	if firstTokenTime.Seconds() <= 0 {
		firstTokenTime = totalDuration
	}

	usage := llms.Usage{
		TotalTime:               totalDuration,
		FirstTokenTime:          firstTokenTime,
		AverageTokensPerSecond:  float64(res.CompletionTokens) / time.Since(startTime).Seconds(),
		PromptTokens:            res.PromptTokens,
		CompletionTokens:        res.CompletionTokens,
		TotalTokens:             res.TotalTokens,
		PromptTokensDetails:     llms.PromptTokensDetail{},
		CompletionTokensDetails: llms.CompletionTokensDetails{},
	}

	if res.PromptTokensDetails != nil {
		usage.PromptTokensDetails = llms.PromptTokensDetail{
			CachedTokens: res.PromptTokensDetails.CachedTokens,
		}
	}

	if res.CompletionTokensDetails != nil {
		usage.CompletionTokensDetails = llms.CompletionTokensDetails{
			ReasoningTokens: res.CompletionTokensDetails.ReasoningTokens,
		}
	}

	return usage
}
