// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package llms

import "testing"

func TestGetBufferString(t *testing.T) {
	t.Parallel()
	testCases := []struct {
		name        string
		messages    []ChatMessage
		humanPrefix string
		aiPrefix    string
		expected    string
		expectError bool
	}{
		{
			name:        "No messages",
			messages:    []ChatMessage{},
			humanPrefix: "Human",
			aiPrefix:    "AI",
			expected:    "",
			expectError: false,
		},
		{
			name: "Mixed messages",
			messages: []ChatMessage{
				SystemChatMessage{Content: "Please be polite."},
				HumanChatMessage{Content: "Hello, how are you?"},
				AIChatMessage{Content: "I'm doing great!"},
				GenericChatMessage{Role: "Moderator", Content: "Keep the conversation on topic."},
			},
			humanPrefix: "Human",
			aiPrefix:    "AI",
			expected:    "system: Please be polite.\nHuman: Hello, how are you?\nAI: I'm doing great!\nModerator: Keep the conversation on topic.", //nolint:lll
			expectError: false,
		},
		{
			name: "Unsupported message type",
			messages: []ChatMessage{
				unsupportedChatMessage{},
			},
			humanPrefix: "Human",
			aiPrefix:    "AI",
			expected:    "",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := GetBufferString(tc.messages, tc.humanPrefix, tc.aiPrefix)
			if (err != nil) != tc.expectError {
				t.Fatalf("expected error: %v, got: %v", tc.expectError, err)
			}

			if result != tc.expected {
				t.Errorf("expected: %q, got: %q", tc.expected, result)
			}
		})
	}
}

type unsupportedChatMessage struct{}

func (m unsupportedChatMessage) GetType() ChatMessageType { return "unsupported" }
func (m unsupportedChatMessage) GetContent() string       { return "Unsupported message" }
