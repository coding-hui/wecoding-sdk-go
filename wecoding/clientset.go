// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package wecoding

import (
	"github.com/coding-hui/wecoding-sdk-go/rest"
	"github.com/coding-hui/wecoding-sdk-go/wecoding/iam"
)

// Interface defines method used to return client interface used by coding-hui organization.
type Interface interface {
	Iam() iam.IamInterface
}

// Clientset contains the clients for groups. Each group has exactly one
// version included in a Clientset.
type Clientset struct {
	iam *iam.IamClient
}

var _ Interface = &Clientset{}

// Iam retrieves the IamClient.
func (c *Clientset) Iam() iam.IamInterface {
	return c.iam
}

// NewForConfig creates a new Clientset for the given config.
// If config's RateLimiter is not set and QPS and Burst are acceptable,
// NewForConfig will generate a rate-limiter in configShallowCopy.
func NewForConfig(c *rest.Config) (*Clientset, error) {
	configShallowCopy := *c

	var cs Clientset

	var err error

	cs.iam, err = iam.NewForConfig(&configShallowCopy)
	if err != nil {
		return nil, err
	}

	return &cs, nil
}

// NewForConfigOrDie creates a new Clientset for the given config and
// panics if there is an error in the config.
func NewForConfigOrDie(c *rest.Config) *Clientset {
	var cs Clientset
	cs.iam = iam.NewForConfigOrDie(c)
	return &cs
}

// New creates a new Clientset for the given RESTClient.
func New(c rest.Interface) *Clientset {
	var cs Clientset
	cs.iam = iam.New(c)
	return &cs
}
