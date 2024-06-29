// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package llms

// PromptValue is the interface that all prompt values must implement.
type PromptValue interface {
	String() string
	Messages() []ChatMessage
}
