// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package rest

import (
	"os"
	"time"
)

const (
	envAccessKeyId     = "IAM_ACCESS_KEY_ID"
	envSecretAccessKey = "IAM_SECRET_ACCESS_KEY"
	envEndpoint        = "IAM_ENDPOINT"
	envAPIServer       = "IAM_API_SERVER_ENDPOINT"
	envAuthzEndpoint   = "IAM_AUTHZ_ENDPOINT"
)

// ConfigOption is a functional option for Config.
type ConfigOption func(*Config)

// NewConfig creates a Config with default values and applies options.
func NewConfig(opts ...ConfigOption) *Config {
	c := &Config{
		Timeout: 30 * time.Second,
	}
	// Apply environment variable fallbacks first
	applyEnvDefaults(c)
	// Then apply user-provided options (override env defaults)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// applyEnvDefaults applies environment variable fallbacks to the config.
func applyEnvDefaults(c *Config) {
	if c.Host == "" {
		if v := os.Getenv(envEndpoint); v != "" {
			c.Host = v
		}
	}
	if c.APIServerEndpoint == "" {
		if v := os.Getenv(envAPIServer); v != "" {
			c.APIServerEndpoint = v
		}
	}
	if c.AuthzEndpoint == "" {
		if v := os.Getenv(envAuthzEndpoint); v != "" {
			c.AuthzEndpoint = v
		}
	}
	if c.Username == "" && c.AccessKeyId == "" {
		if v := os.Getenv(envAccessKeyId); v != "" {
			c.AccessKeyId = v
		}
	}
	if c.Password == "" && c.SecretAccessKey == "" {
		if v := os.Getenv(envSecretAccessKey); v != "" {
			c.SecretAccessKey = v
		}
	}
}

// WithHost sets the host (legacy, use WithEndpoint instead).
func WithHost(addr string) ConfigOption {
	return func(c *Config) {
		c.Host = addr
	}
}

// WithEndpoint sets the primary endpoint.
func WithEndpoint(addr string) ConfigOption {
	return func(c *Config) {
		c.Host = addr
	}
}

// WithAPIServerEndpoint sets the API server endpoint.
func WithAPIServerEndpoint(addr string) ConfigOption {
	return func(c *Config) {
		c.APIServerEndpoint = addr
	}
}

// WithAuthzEndpoint sets the Authz server endpoint.
func WithAuthzEndpoint(addr string) ConfigOption {
	return func(c *Config) {
		c.AuthzEndpoint = addr
	}
}

// WithCredentials sets the access key ID and secret access key.
func WithCredentials(accessKeyId, secretAccessKey string) ConfigOption {
	return func(c *Config) {
		c.AccessKeyId = accessKeyId
		c.SecretAccessKey = secretAccessKey
	}
}

// WithAccessKeyId sets the access key ID.
func WithAccessKeyId(id string) ConfigOption {
	return func(c *Config) {
		c.AccessKeyId = id
	}
}

// WithSecretAccessKey sets the secret access key.
func WithSecretAccessKey(key string) ConfigOption {
	return func(c *Config) {
		c.SecretAccessKey = key
	}
}

// WithBearerToken sets the bearer token.
func WithBearerToken(token string) ConfigOption {
	return func(c *Config) {
		c.BearerToken = token
	}
}

// WithTimeout sets the timeout.
func WithTimeout(d time.Duration) ConfigOption {
	return func(c *Config) {
		c.Timeout = d
	}
}

// WithCACertificate sets the CA certificate data.
func WithCACertificate(caCert []byte) ConfigOption {
	return func(c *Config) {
		c.TLSClientConfig.CAData = caCert
	}
}
