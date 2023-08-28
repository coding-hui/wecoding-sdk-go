// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package rest

// CommonResponse Http API common response.
type CommonResponse struct {
	// Success request is successful
	Success bool `json:"success"`

	// Code defines the business error code.
	Code int `json:"code"`

	// Msg contains the detail of this message.
	// This message is suitable to be exposed to external
	Msg string `json:"msg"`

	// Data return data object
	Data interface{} `json:"data,omitempty"`

	// Reference returns the reference document which maybe useful to solve this error.
	Reference string `json:"reference,omitempty"`
}

// PageInfo Http API common page info.
type PageInfo struct {
	// List all records
	List interface{} `json:"list"`
	// Total all count
	Total int64 `json:"total"`
}
