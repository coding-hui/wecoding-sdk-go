package volcengine

import (
	"context"
	"errors"
	"fmt"
	"io"

	sdk "github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	chat "github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
	"github.com/volcengine/volcengine-go-sdk/volcengine"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/callbacks"
	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
)

var (
	ErrEmptyResponse              = errors.New("no response")
	ErrMissingApiTokenOrAccessKey = errors.New("missing the volcano api token or access key, set it in the VOLC_ACCESSKEY environment variable or VOLC_SECRETKEY environment variable") //nolint:lll
	ErrMissingModel               = errors.New("model needs to be provided when using Volcano API")
	ErrMissingEmbeddingModel      = errors.New("embeddings model needs to be provided when using Volcano API")
	ErrUnexpectedResponseLength   = errors.New("unexpected length of response")
)

// ResponseFormatJSON is the JSON response format.
var ResponseFormatJSON = &chat.ResponseFormat{Type: chat.ResponseFormatJsonObject} //nolint:gochecknoglobals

type Option = sdk.ConfigOption

type Model struct {
	client *sdk.Client

	CallbacksHandler callbacks.Handler
}

var _ llms.Model = (*Model)(nil)

func NewClientWithApiKey(apiKey string, opts ...Option) (*Model, error) {
	return NewClientWithConfig(apiKey, "", "", opts...)
}

func NewClientWithAkSk(ak, sk string, opts ...Option) (*Model, error) {
	return NewClientWithConfig("", ak, sk, opts...)
}

func NewClientWithConfig(apiKey, ak, sk string, opts ...Option) (*Model, error) {
	if len(apiKey) == 0 && len(sk) == 0 && len(ak) == 0 {
		return nil, ErrMissingApiTokenOrAccessKey
	}

	var client *sdk.Client
	if len(apiKey) > 0 {
		client = sdk.NewClientWithApiKey(apiKey, opts...)
	} else {
		client = sdk.NewClientWithAkSk(ak, sk, opts...)
	}

	return &Model{client: client}, nil
}

// GenerateContent implements the Model interface.
func (o *Model) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) { //nolint: lll, cyclop, goerr113, funlen
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentStart(ctx, messages)
	}

	opts := llms.CallOptions{}
	for _, opt := range options {
		opt(&opts)
	}

	chatMsgs := make([]*chat.ChatCompletionMessage, 0, len(messages))
	for _, mc := range messages {
		content, multiContent := messagePartsFromParts(mc.Parts)
		msg := &chat.ChatCompletionMessage{
			Content: &chat.ChatCompletionMessageContent{StringValue: volcengine.String(content)},
		}
		if opts.SupportMultiContent {
			msg = &chat.ChatCompletionMessage{
				Content: &chat.ChatCompletionMessageContent{ListValue: multiContent},
			}
		}
		switch mc.Role {
		case llms.ChatMessageTypeSystem:
			msg.Role = chat.ChatMessageRoleSystem
		case llms.ChatMessageTypeAI:
			msg.Role = chat.ChatMessageRoleAssistant
		case llms.ChatMessageTypeHuman:
			msg.Role = chat.ChatMessageRoleUser
		case llms.ChatMessageTypeGeneric:
			msg.Role = chat.ChatMessageRoleUser
		default:
			return nil, fmt.Errorf("role %v not supported", mc.Role)
		}

		chatMsgs = append(chatMsgs, msg)
	}

	req := chat.ChatCompletionRequest{
		Model:            opts.Model,
		Stop:             opts.StopWords,
		Messages:         chatMsgs,
		Temperature:      float32(opts.Temperature),
		MaxTokens:        opts.MaxTokens,
		N:                opts.N,
		FrequencyPenalty: float32(opts.FrequencyPenalty),
		PresencePenalty:  float32(opts.PresencePenalty),

		ToolChoice: opts.ToolChoice,
	}
	if opts.JSONMode {
		req.ResponseFormat = ResponseFormatJSON
	}

	if opts.StreamingFunc != nil {
		return o.parseStreamingChatResponse(ctx, req, opts.StreamingFunc)
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
			Content:          volcengine.StringValue(c.Message.Content.StringValue),
			ReasoningContent: volcengine.StringValue(c.Message.ReasoningContent),
			StopReason:       fmt.Sprint(c.FinishReason),
			GenerationInfo: map[string]any{
				"CompletionTokens": result.Usage.CompletionTokens,
				"PromptTokens":     result.Usage.PromptTokens,
				"TotalTokens":      result.Usage.TotalTokens,
			},
		}
	}

	response := &llms.ContentResponse{Choices: choices}
	if o.CallbacksHandler != nil {
		o.CallbacksHandler.HandleLLMGenerateContentEnd(ctx, response)
	}

	return response, nil
}

func (o *Model) parseStreamingChatResponse(
	ctx context.Context,
	req chat.ChatCompletionRequest,
	streamingFunc func(ctx context.Context, chunk []byte) error,
) (*llms.ContentResponse, error) {
	req.Stream = true
	req.StreamOptions = &chat.StreamOptions{
		IncludeUsage: true,
	}

	stream, err := o.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		if o.CallbacksHandler != nil {
			o.CallbacksHandler.HandleLLMError(ctx, err)
		}
		return nil, err
	}

	responseChan := make(chan chat.ChatCompletionStreamResponse)
	go func() {
		defer close(responseChan)
		for {
			chunk, err := stream.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					return
				}
				chunk.Choices = append(chunk.Choices, &chat.ChatCompletionStreamChoice{
					Delta: chat.ChatCompletionStreamChoiceDelta{
						Content: fmt.Errorf("error decoding streaming response: %w", err).Error(),
					},
				})
				responseChan <- chunk
				return
			}
			responseChan <- chunk
		}
	}()

	response, err := o.combineStreamingChatResponse(ctx, responseChan, streamingFunc)
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
	responseChan chan chat.ChatCompletionStreamResponse,
	streamingFunc func(ctx context.Context, chunk []byte) error,
) (*llms.ContentResponse, error) {
	response := llms.ContentResponse{
		Choices: []*llms.ContentChoice{
			{},
		},
	}

	for streamResponse := range responseChan {
		if streamResponse.Usage != nil {
			response.Choices[0].GenerationInfo = map[string]any{
				"CompletionTokens": streamResponse.Usage.CompletionTokens,
				"PromptTokens":     streamResponse.Usage.PromptTokens,
				"TotalTokens":      streamResponse.Usage.TotalTokens,
			}
		}

		if len(streamResponse.Choices) == 0 {
			continue
		}

		choice := streamResponse.Choices[0]
		response.Choices[0].Content += choice.Delta.Content
		response.Choices[0].ReasoningContent += volcengine.StringValue(choice.Delta.ReasoningContent)
		response.Choices[0].StopReason = fmt.Sprint(choice.FinishReason)

		chunk := []byte(choice.Delta.Content)
		if choice.Delta.ReasoningContent != nil {
			chunk = []byte(volcengine.StringValue(choice.Delta.ReasoningContent))
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

func messagePartsFromParts(parts []llms.ContentPart) (string, []*chat.ChatCompletionMessageContentPart) {
	var content []*chat.ChatCompletionMessageContentPart
	fullContent := ""
	for _, part := range parts {
		switch p := part.(type) {
		case llms.TextContent:
			content = append(content, &chat.ChatCompletionMessageContentPart{
				Type:     chat.ChatCompletionMessageContentPartTypeText,
				Text:     p.Text,
				ImageURL: nil,
			})
			fullContent += p.Text
		case llms.ImageURLContent:
			content = append(content, &chat.ChatCompletionMessageContentPart{
				Type: chat.ChatCompletionMessageContentPartTypeImageURL,
				ImageURL: &chat.ChatMessageImageURL{
					URL:    p.URL,
					Detail: chat.ImageURLDetail(p.Detail),
				},
			})
		}
	}
	return fullContent, content
}
