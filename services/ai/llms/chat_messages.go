// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package llms

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/coding-hui/common/errors"
)

// ErrUnexpectedChatMessageType is returned when a chat message is of an
// unexpected type.
var ErrUnexpectedChatMessageType = errors.New("unexpected chat message type")

// ChatMessageType is the type of chat message.
type ChatMessageType string

const (
	// ChatMessageTypeAI is a message sent by an AI.
	ChatMessageTypeAI ChatMessageType = "ai"
	// ChatMessageTypeHuman is a message sent by a human.
	ChatMessageTypeHuman ChatMessageType = "human"
	// ChatMessageTypeSystem is a message sent by the system.
	ChatMessageTypeSystem ChatMessageType = "system"
	// ChatMessageTypeGeneric is a message sent by a generic user.
	ChatMessageTypeGeneric ChatMessageType = "generic"
	// ChatMessageTypeFunction is a message sent by a function.
	ChatMessageTypeFunction ChatMessageType = "function"
	// ChatMessageTypeTool is a message sent by a tool.
	ChatMessageTypeTool ChatMessageType = "tool"
)

// ChatMessage represents a message in a chat.
type ChatMessage interface {
	// GetType gets the type of the message.
	GetType() ChatMessageType
	// GetContent gets the content of the message.
	GetContent() string
}

// Named is an interface for objects that have a name.
type Named interface {
	GetName() string
}

// Reasoning is an interface for objects that have a reasoning content.
type Reasoning interface {
	GetReasoningContent() string
}

// Statically assert that the types implement the interface.
var (
	_ ChatMessage = AIChatMessage{}
	_ ChatMessage = HumanChatMessage{}
	_ ChatMessage = SystemChatMessage{}
	_ ChatMessage = GenericChatMessage{}
	_ ChatMessage = ToolChatMessage{}
)

// AIChatMessage is a message sent by an AI.
type AIChatMessage struct {
	// Content is the content of the message.
	Content string `json:"content,omitempty"`

	// ReasoningContent is the reasoning content of a response
	ReasoningContent string `json:"reasoning_content,omitempty"`

	// FunctionCall represents the model choosing to call a function.
	FunctionCall *FunctionCall `json:"function_call,omitempty"`

	// ToolCalls represents the model choosing to call tools.
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

func (m AIChatMessage) GetType() ChatMessageType {
	return ChatMessageTypeAI
}

func (m AIChatMessage) GetContent() string {
	return m.Content
}

func (m AIChatMessage) GetFunctionCall() *FunctionCall {
	return m.FunctionCall
}

func (m AIChatMessage) GetReasoningContent() string {
	return m.ReasoningContent
}

// HumanChatMessage is a message sent by a human.
type HumanChatMessage struct {
	Content string
}

func (m HumanChatMessage) GetType() ChatMessageType {
	return ChatMessageTypeHuman
}

func (m HumanChatMessage) GetContent() string {
	return m.Content
}

// SystemChatMessage is a chat message representing information that should be instructions to the AI system.
type SystemChatMessage struct {
	Content string
}

func (m SystemChatMessage) GetType() ChatMessageType {
	return ChatMessageTypeSystem
}

func (m SystemChatMessage) GetContent() string {
	return m.Content
}

// GenericChatMessage is a chat message with an arbitrary speaker.
type GenericChatMessage struct {
	Content string
	Role    string
	Name    string
}

func (m GenericChatMessage) GetType() ChatMessageType {
	return ChatMessageTypeGeneric
}

func (m GenericChatMessage) GetContent() string {
	return m.Content
}

func (m GenericChatMessage) GetName() string {
	return m.Name
}

// ToolChatMessage is a chat message representing the result of a tool call.
type ToolChatMessage struct {
	// ID is the ID of the tool call.
	ID string `json:"tool_call_id"`
	// Content is the content of the tool message.
	Content string `json:"content"`
}

func (m ToolChatMessage) GetType() ChatMessageType {
	return ChatMessageTypeTool
}

func (m ToolChatMessage) GetContent() string {
	return m.Content
}

func (m ToolChatMessage) GetID() string {
	return m.ID
}

// GetBufferString gets the buffer string of messages.
func GetBufferString(messages []ChatMessage, systemPrefix, humanPrefix string, aiPrefix string) (string, error) {
	var result []string
	for _, m := range messages {
		role, err := getMessageRole(m, systemPrefix, humanPrefix, aiPrefix)
		if err != nil {
			return "", err
		}
		msg := fmt.Sprintf("%s: %s", role, m.GetContent())
		if m, ok := m.(AIChatMessage); ok && m.FunctionCall != nil {
			j, err := json.Marshal(m.FunctionCall)
			if err != nil {
				return "", err
			}
			msg = fmt.Sprintf("%s %s", msg, string(j))
		}
		result = append(result, msg)
	}
	return strings.Join(result, "\n"), nil
}

func getMessageRole(m ChatMessage, systemPrefix, humanPrefix, aiPrefix string) (string, error) {
	var role string
	switch m.GetType() {
	case ChatMessageTypeHuman:
		role = humanPrefix
	case ChatMessageTypeAI:
		role = aiPrefix
	case ChatMessageTypeSystem:
		role = systemPrefix
	case ChatMessageTypeGeneric:
		cgm, ok := m.(GenericChatMessage)
		if !ok {
			return "", fmt.Errorf("%w -%+v", ErrUnexpectedChatMessageType, m)
		}
		role = cgm.Role
	case ChatMessageTypeFunction:
		role = "function"
	case ChatMessageTypeTool:
		role = "tool"
	default:
		return "", ErrUnexpectedChatMessageType
	}
	return role, nil
}

type ChatMessageModelData struct {
	Content          string `bson:"content" json:"content"`
	ReasoningContent string `bson:"reasoning_content" json:"reasoning_content"`
	Type             string `bson:"type"    json:"type"`
}

type ChatMessageModel struct {
	Type string               `bson:"type" json:"type"`
	Data ChatMessageModelData `bson:"data" json:"data"`
}

func (c ChatMessageModel) ToChatMessage() ChatMessage {
	switch c.Type {
	case string(ChatMessageTypeAI):
		return AIChatMessage{Content: c.Data.Content, ReasoningContent: c.Data.ReasoningContent}
	case string(ChatMessageTypeHuman):
		return HumanChatMessage{Content: c.Data.Content}
	case string(ChatMessageTypeSystem):
		return SystemChatMessage{Content: c.Data.Content}
	case string(ChatMessageTypeGeneric):
		return GenericChatMessage{Content: c.Data.Content}
	case string(ChatMessageTypeFunction):
		return ToolChatMessage{Content: c.Data.Content}
	case string(ChatMessageTypeTool):
		return ToolChatMessage{Content: c.Data.Content}
	default:
		log.Println("convert to chat message failed with invalid message type", "type", c.Type)
		return nil
	}
}

// ConvertChatMessageToModel Convert a ChatMessage to a ChatMessageModel.
func ConvertChatMessageToModel(m ChatMessage) ChatMessageModel {
	return ChatMessageModel{
		Type: string(m.GetType()),
		Data: ChatMessageModelData{
			Type:    string(m.GetType()),
			Content: m.GetContent(),
		},
	}
}
