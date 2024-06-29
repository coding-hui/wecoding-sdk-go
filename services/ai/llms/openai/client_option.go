// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package openai

import (
	"net/http"

	sdk "github.com/sashabaranov/go-openai"
)

const (
	tokenEnvVarName        = "OPENAI_API_KEY"      //nolint:gosec
	modelEnvVarName        = "OPENAI_MODEL"        //nolint:gosec
	baseURLEnvVarName      = "OPENAI_BASE_URL"     //nolint:gosec
	baseAPIBaseEnvVarName  = "OPENAI_API_BASE"     //nolint:gosec
	organizationEnvVarName = "OPENAI_ORGANIZATION" //nolint:gosec
)

type APIType = sdk.APIType

const (
	APITypeOpenAI  APIType = sdk.APITypeOpenAI
	APITypeAzure           = sdk.APITypeAzure
	APITypeAzureAD         = sdk.APITypeAzureAD
)

const (
	DefaultAPIVersion = "2023-05-15"
)

type clientOptions struct {
	token        string
	model        string
	baseURL      string
	organization string
	apiType      APIType
	httpClient   *http.Client

	responseFormat *ResponseFormat

	// required when APIType is APITypeAzure or APITypeAzureAD
	apiVersion     string
	embeddingModel string
}

// Option is a functional option for the OpenAI client.
type Option func(*clientOptions)

// ResponseFormat is the response format for the OpenAI client.
type ResponseFormat = sdk.ReponseFormat

// ChatCompletionResponseFormat is the chat response format for the OpenAI client.
type ChatCompletionResponseFormat = sdk.ChatCompletionResponseFormat

// ResponseFormatJSON is the JSON response format.
var ResponseFormatJSON = &ResponseFormat{Type: "json_object"} //nolint:gochecknoglobals

// ChatCompletionResponseFormatJSON is the JSON response format.
var ChatCompletionResponseFormatJSON = &ChatCompletionResponseFormat{Type: sdk.ChatCompletionResponseFormatTypeJSONObject} //nolint:gochecknoglobals

// WithToken passes the OpenAI API token to the client. If not set, the token
// is read from the OPENAI_API_KEY environment variable.
func WithToken(token string) Option {
	return func(opts *clientOptions) {
		opts.token = token
	}
}

// WithModel passes the OpenAI model to the client. If not set, the model
// is read from the OPENAI_MODEL environment variable.
// Required when ApiType is Azure.
func WithModel(model string) Option {
	return func(opts *clientOptions) {
		opts.model = model
	}
}

// WithEmbeddingModel passes the OpenAI model to the client. Required when ApiType is Azure.
func WithEmbeddingModel(embeddingModel string) Option {
	return func(opts *clientOptions) {
		opts.embeddingModel = embeddingModel
	}
}

// WithBaseURL passes the OpenAI base url to the client. If not set, the base url
// is read from the OPENAI_BASE_URL environment variable. If still not set in ENV
// VAR OPENAI_BASE_URL, then the default value is https://api.openai.com/v1 is used.
func WithBaseURL(baseURL string) Option {
	return func(opts *clientOptions) {
		opts.baseURL = baseURL
	}
}

// WithOrganization passes the OpenAI organization to the client. If not set, the
// organization is read from the OPENAI_ORGANIZATION.
func WithOrganization(organization string) Option {
	return func(opts *clientOptions) {
		opts.organization = organization
	}
}

// WithAPIType passes the api type to the client. If not set, the default value
// is APITypeOpenAI.
func WithAPIType(apiType APIType) Option {
	return func(opts *clientOptions) {
		opts.apiType = apiType
	}
}

// WithAPIVersion passes the api version to the client. If not set, the default value
// is DefaultAPIVersion.
func WithAPIVersion(apiVersion string) Option {
	return func(opts *clientOptions) {
		opts.apiVersion = apiVersion
	}
}

// WithResponseFormat allows setting a custom response format.
func WithResponseFormat(responseFormat *ResponseFormat) Option {
	return func(opts *clientOptions) {
		opts.responseFormat = responseFormat
	}
}

// WithHTTPClient allows setting a custom HTTP client. If not set, the default value
// is http.DefaultClient.
func WithHTTPClient(client *http.Client) Option {
	return func(opts *clientOptions) {
		opts.httpClient = client
	}
}
