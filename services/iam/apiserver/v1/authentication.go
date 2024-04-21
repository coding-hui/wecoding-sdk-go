// Copyright (c) 2024 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package v1

import (
	"context"
	"fmt"

	v1 "github.com/coding-hui/iam/pkg/api/apiserver/v1"

	"github.com/coding-hui/wecoding-sdk-go/rest"
)

type AuthenticationGetter interface {
	Authentication() AuthenticationInterface
}

type AuthenticationInterface interface {
	Login(ctx context.Context, username, password string) (*v1.AuthenticateResponse, error)
	Authenticate(ctx context.Context, loginReq v1.AuthenticateRequest) (*v1.AuthenticateResponse, error)
	RefreshToken(ctx context.Context, refreshToken ...string) (*v1.RefreshTokenResponse, error)
	UserInfo(ctx context.Context, accessToken ...string) (*v1.DetailUserResponse, error)
	AuthenticationExpansion
}

type AuthenticationExpansion interface{}

type authentication struct {
	client rest.Interface
}

func newAuthentication(c *APIV1Client) *authentication {
	return &authentication{
		client: c.RESTClient(),
	}
}

func (a *authentication) Authenticate(ctx context.Context, loginReq v1.AuthenticateRequest) (*v1.AuthenticateResponse, error) {
	result := &v1.AuthenticateResponse{}
	err := a.client.Post().
		Suffix("/login").
		Body(loginReq).
		Do(ctx).
		Into(result)
	return result, err
}

func (a *authentication) Login(ctx context.Context, username, password string) (*v1.AuthenticateResponse, error) {
	result := &v1.AuthenticateResponse{}
	err := a.client.Post().
		Suffix("/login").
		Body(v1.AuthenticateRequest{
			Username: username,
			Password: password,
		}).
		Do(ctx).
		Into(result)
	return result, err
}

func (a *authentication) RefreshToken(ctx context.Context, refreshToken ...string) (*v1.RefreshTokenResponse, error) {
	request := a.client.Get().Suffix("/auth/refresh-token")
	if len(refreshToken) >= 0 {
		request.SetHeader("RefreshToken", refreshToken[0])
	}
	result := &v1.RefreshTokenResponse{}
	err := request.Do(ctx).Into(result)
	return result, err
}

func (a *authentication) UserInfo(ctx context.Context, accessToken ...string) (*v1.DetailUserResponse, error) {
	request := a.client.Get().Suffix("/auth/user-info")
	if len(accessToken) >= 0 {
		request.SetHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken[0]))
	}
	result := &v1.DetailUserResponse{}
	err := request.Do(ctx).Into(result)
	return result, err
}
