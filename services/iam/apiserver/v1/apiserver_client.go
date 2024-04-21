// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package v1

import (
	"github.com/coding-hui/common/runtime"
	v1 "github.com/coding-hui/iam/pkg/api/apiserver/v1"

	"github.com/coding-hui/wecoding-sdk-go/rest"
)

// APIV1Interface has methods to work with iam resources.
type APIV1Interface interface {
	RESTClient() rest.Interface
	UsersGetter
	AuthenticationGetter
}

// APIV1Client is used to interact with features provided by the group.
type APIV1Client struct {
	restClient rest.Interface
}

// Users create and return user rest client.
func (c *APIV1Client) Users() UserInterface {
	return newUsers(c)
}

// Authentication create and return user rest client.
func (c *APIV1Client) Authentication() AuthenticationInterface {
	return newAuthentication(c)
}

// NewForConfig creates a new APIV1Client for the given config.
func NewForConfig(c *rest.Config) (*APIV1Client, error) {
	config := *c
	setConfigDefaults(&config)

	client, err := rest.RESTClientFor(&config)
	if err != nil {
		return nil, err
	}

	return &APIV1Client{client}, nil
}

// NewForConfigOrDie creates a new APIV1Client for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *APIV1Client {
	client, err := NewForConfig(c)
	if err != nil {
		panic(err)
	}

	return client
}

// New creates a new APIV1Client for the given RESTClient.
func New(c rest.Interface) *APIV1Client {
	return &APIV1Client{c}
}

func setConfigDefaults(config *rest.Config) {
	gv := v1.SchemeGroupVersion
	config.GroupVersion = &gv
	config.APIPath = ""
	config.Negotiator = runtime.NewSimpleClientNegotiator()

	if config.UserAgent == "" {
		config.UserAgent = rest.DefaultUserAgent()
	}
}

// RESTClient returns a RESTClient that is used to communicate
// with API server by this client implementation.
func (c *APIV1Client) RESTClient() rest.Interface {
	if c == nil {
		return nil
	}

	return c.restClient
}
