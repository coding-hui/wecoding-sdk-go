// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package clientcmd

import (
	"net/url"
	"time"

	restclient "github.com/coding-hui/wecoding-sdk-go/rest"
)

// Server contains information about how to communicate with a iam api server.
type Server struct {
	LocationOfOrigin string
	Timeout          time.Duration `yaml:"timeout,omitempty"                    mapstructure:"timeout,omitempty"`
	MaxRetries       int           `yaml:"maxRetries,omitempty"                mapstructure:"maxRetries,omitempty"`
	RetryInterval    time.Duration `yaml:"retryInterval,omitempty"             mapstructure:"retryInterval,omitempty"`
	// Endpoint is the primary endpoint for the server.
	Endpoint string `yaml:"endpoint,omitempty"                    mapstructure:"endpoint,omitempty"`
	// APIServerEndpoint is the endpoint for the API server.
	APIServerEndpoint string `yaml:"apiServerEndpoint,omitempty"               mapstructure:"apiServerEndpoint,omitempty"`
	// AuthzEndpoint is the endpoint for the Authz server.
	AuthzEndpoint string `yaml:"authzEndpoint,omitempty"               mapstructure:"authzEndpoint,omitempty"`

	// TLSServerName is used to check server certificate. If TLSServerName is empty, the hostname used to contact the
	// server is used.
	// +optional
	TLSServerName string `yaml:"tlsServerName,omitempty"            mapstructure:"tlsServerName,omitempty"`
	// InsecureSkipTLSVerify skips the validity check for the server's certificate. This will make your HTTPS
	// connections insecure.
	// +optional
	InsecureSkipTLSVerify bool `yaml:"insecureSkipTLSVerify,omitempty"   mapstructure:"insecureSkipTLSVerify,omitempty"`
	// CertificateAuthority is the path to a cert file for the certificate authority.
	// +optional
	CertificateAuthority string `yaml:"certificateAuthority,omitempty"      mapstructure:"certificateAuthority,omitempty"`
	// CertificateAuthorityData contains PEM-encoded certificate authority certificates.
	// Overrides CertificateAuthority
	// +optional
	CertificateAuthorityData string `yaml:"certificateAuthorityData,omitempty" mapstructure:"certificateAuthorityData,omitempty"`
}

// Credentials contains information that describes identity information.
// This is use to tell the iam cluster who you are.
type Credentials struct {
	LocationOfOrigin  string
	ClientCertificate string `yaml:"clientCertificate,omitempty"         mapstructure:"clientCertificate,omitempty"`
	// ClientCertificateData contains PEM-encoded data from a client cert file for TLS. Overrides ClientCertificate
	// +optional
	ClientCertificateData string `yaml:"clientCertificateData,omitempty"   mapstructure:"clientCertificateData,omitempty"`
	// ClientKey is the path to a client key file for TLS.
	// +optional
	ClientKey string `yaml:"clientKey,omitempty"              mapstructure:"clientKey,omitempty"`
	// ClientKeyData contains PEM-encoded data from a client key file for TLS. Overrides ClientKey
	// +optional
	ClientKeyData string `yaml:"clientKeyData,omitempty"         mapstructure:"clientKeyData,omitempty"`
	// Token is the bearer token for authentication to the iam cluster.
	// +optional
	Token string `yaml:"token,omitempty"                   mapstructure:"token,omitempty"`

	Username string `yaml:"username,omitempty" mapstructure:"username,omitempty"`
	Password string `yaml:"password,omitempty" mapstructure:"password,omitempty"`

	// AccessKeyId is the AWS-style access key ID.
	AccessKeyId string `yaml:"accessKeyId,omitempty" mapstructure:"accessKeyId,omitempty"`
	// SecretAccessKey is the AWS-style secret access key.
	SecretAccessKey string `yaml:"secretAccessKey,omitempty" mapstructure:"secretAccessKey,omitempty"`
}

// AuthInfo is an alias for Credentials for backward compatibility.
type AuthInfo = Credentials

// Config defines a config struct used by marmotedu-sdk-go.
type Config struct {
	APIVersion  string       `yaml:"apiVersion,omitempty" mapstructure:"apiVersion,omitempty"`
	Credentials *Credentials `yaml:"credentials,omitempty" mapstructure:"credentials,omitempty"`
	Server      *Server      `yaml:"server,omitempty"     mapstructure:"server,omitempty"`
}

// NewConfig is a convenience function that returns a new Config object with non-nil maps.
func NewConfig() *Config {
	return &Config{
		Server:      &Server{},
		Credentials: &Credentials{},
	}
}

// ClientConfig is used to make it easy to get an api server client.
type ClientConfig interface {
	// ClientConfig returns a complete client config
	ClientConfig() (*restclient.Config, error)
}

// DirectClientConfig wrap for Config.
type DirectClientConfig struct {
	config Config
}

// NewClientConfigFromConfig takes your Config and gives you back a ClientConfig.
func NewClientConfigFromConfig(config *Config) ClientConfig {
	return &DirectClientConfig{*config}
}

// NewClientConfigFromBytes takes your iamconfig and gives you back a ClientConfig.
func NewClientConfigFromBytes(configBytes []byte) (ClientConfig, error) {
	config, err := Load(configBytes)
	if err != nil {
		return nil, err
	}

	return &DirectClientConfig{*config}, nil
}

// RESTConfigFromIAMConfig is a convenience method to give back a restconfig from your iamconfig bytes.
// For programmatic access, this is what you want 80% of the time.
func RESTConfigFromIAMConfig(configBytes []byte) (*restclient.Config, error) {
	clientConfig, err := NewClientConfigFromBytes(configBytes)
	if err != nil {
		return nil, err
	}

	return clientConfig.ClientConfig()
}

// ClientConfig implements ClientConfig.
func (config *DirectClientConfig) ClientConfig() (*restclient.Config, error) {
	user := config.getAuthInfo()
	server := config.getServer()

	if err := config.ConfirmUsable(); err != nil {
		return nil, err
	}

	clientConfig := &restclient.Config{
		BearerToken:        user.Token,
		Username:           user.Username,
		Password:           user.Password,
		AccessKeyId:        user.AccessKeyId,
		SecretAccessKey:    user.SecretAccessKey,
		Host:               server.Endpoint,
		Timeout:       server.Timeout,
		MaxRetries:    server.MaxRetries,
		RetryInterval: server.RetryInterval,
		APIServerEndpoint: server.APIServerEndpoint,
		AuthzEndpoint:    server.AuthzEndpoint,
		TLSClientConfig: restclient.TLSClientConfig{
			Insecure:   server.InsecureSkipTLSVerify,
			ServerName: server.TLSServerName,
			CertFile:   user.ClientCertificate,
			KeyFile:    user.ClientKey,
			CertData:   []byte(user.ClientCertificateData),
			KeyData:    []byte(user.ClientKeyData),
			CAFile:     server.CertificateAuthority,
			CAData:     []byte(server.CertificateAuthorityData),
			// NextProtos []string
		},
	}

	if u, err := url.ParseRequestURI(clientConfig.Host); err == nil && u.Opaque == "" && len(u.Path) > 1 {
		u.RawQuery = ""
		u.Fragment = ""
		clientConfig.Host = u.String()
	}

	return clientConfig, nil
}

// ConfirmUsable looks a particular context and determines if that particular part of
// the config is useable.  There might still be errors in the config, but no errors in the
// sections requested or referenced.  It does not return early so that it can find as many errors as possible.
func (config *DirectClientConfig) ConfirmUsable() error {
	validationErrors := make([]error, 0)

	authInfo := config.getAuthInfo()
	validationErrors = append(validationErrors, validateAuthInfo(authInfo)...)
	server := config.getServer()
	validationErrors = append(validationErrors, validateServerInfo(server)...)
	// when direct client config is specified, and our only error is that no server is defined, we should
	// return a standard "no config" error
	if len(validationErrors) == 1 && validationErrors[0] == ErrEmptyServer {
		return newErrConfigurationInvalid([]error{ErrEmptyConfig})
	}

	return newErrConfigurationInvalid(validationErrors)
}

// getAuthInfo returns the Credentials, or an error if a required auth info is not found.
func (config *DirectClientConfig) getAuthInfo() Credentials {
	return *config.config.Credentials
}

// getServer returns the clientcmdapi.Cluster, or an error if a required cluster is not found.
func (config *DirectClientConfig) getServer() Server {
	return *config.config.Server
}

// BuildConfigFromFlags is a helper function that builds configs from a master
// url or a iamconfig filepath. These are passed in as command line flags for cluster
// components. Warnings should reflect this usage. If neither masterUrl or iamconfigPath
// are passed in we fallback to inClusterConfig. If inClusterConfig fails, we fallback
// to the default config.
func BuildConfigFromFlags(serverURL, iamconfigPath string) (*restclient.Config, error) {
	config, err := LoadFromFile(iamconfigPath)
	if err != nil {
		return nil, err
	}

	if len(serverURL) > 0 {
		config.Server.Endpoint = serverURL
	}

	directClientConfig := &DirectClientConfig{*config}

	return directClientConfig.ClientConfig()
}
