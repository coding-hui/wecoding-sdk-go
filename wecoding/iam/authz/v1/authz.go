// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package v1

import (
	"context"

	v1 "github.com/coding-hui/iam/pkg/api/authzserver/v1"

	"github.com/coding-hui/wecoding-sdk-go/rest"
)

// AuthzGetter has a method to return a AuthzInterface.
// A group's client should implement this interface.
type AuthzGetter interface {
	Authz() AuthzInterface
}

// AuthzInterface has methods to work with Authz resources.
type AuthzInterface interface {
	Authorize(ctx context.Context, request *v1.Request) (*v1.Response, error)
	AuthzExpansion
}

// authz implements AuthzInterface.
type authz struct {
	client rest.Interface
}

// newAuthz returns a Authz.
func newAuthz(c *AuthzV1Client) *authz {
	return &authz{
		client: c.RESTClient(),
	}
}

// Authorize Get takes name of the secret, and returns the corresponding secret object, and an error if there is any.
func (c *authz) Authorize(ctx context.Context, request *v1.Request) (result *v1.Response, err error) {
	result = &v1.Response{}
	err = c.client.Post().
		Resource("authz").
		Body(request).
		Do(ctx).
		Into(result)

	return
}
