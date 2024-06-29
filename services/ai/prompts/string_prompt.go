// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package prompts

import "github.com/coding-hui/wecoding-sdk-go/services/ai/llms"

var _ llms.PromptValue = StringPromptValue("")

// StringPromptValue is a prompt value that is a string.
type StringPromptValue string

func (v StringPromptValue) String() string {
	return string(v)
}

// Messages returns a single-element ChatMessage slice.
func (v StringPromptValue) Messages() []llms.ChatMessage {
	return []llms.ChatMessage{
		llms.HumanChatMessage{Content: string(v)},
	}
}
