package volcengine

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/coding-hui/wecoding-sdk-go/services/ai/llms"
	"github.com/stretchr/testify/assert"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
)

func TestGenerateContent(t *testing.T) {
	ctx := context.Background()
	client, err := NewClientWithApiKey(
		os.Getenv("ARK_API_KEY"),
		arkruntime.WithBaseUrl("https://ark-cn-beijing.bytedance.net/api/v3"),
		// The output time of the reasoning model is relatively long. Please increase the timeout period.
		arkruntime.WithTimeout(30*time.Minute),
	)
	assert.NoError(t, err)
	assert.NotNil(t, client)

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Hello"},
			},
		},
	}

	options := []llms.CallOption{
		llms.WithModel(os.Getenv("ARK_MODEL")),
		llms.WithTemperature(0.5),
		llms.WithMaxTokens(100),
	}

	t.Run("TestGenerateContent", func(t *testing.T) {
		response, err := client.GenerateContent(ctx, messages, options...)
		fmt.Printf("Received response: %s\n", response.Choices[0].Content)
		assert.NoError(t, err)
		assert.NotNil(t, response)
	})

	t.Run("TestGenerateStream", func(t *testing.T) {
		streamingFunc := func(ctx context.Context, chunk []byte) error {
			fmt.Printf("Received chunk: %s\n", chunk)
			return nil
		}

		options = append(options, llms.WithStreamingFunc(streamingFunc))
		response, err := client.GenerateContent(ctx, messages, options...)

		fmt.Printf("Received response: %s\n", response.Choices[0].Content)

		assert.NoError(t, err)
		assert.NotNil(t, response)
	})
}
