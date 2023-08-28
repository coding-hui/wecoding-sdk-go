// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package v1

import (
	"context"

	metav1 "github.com/coding-hui/common/meta/v1"
	v1 "github.com/coding-hui/iam/pkg/api/apiserver/v1"

	"github.com/coding-hui/wecoding-sdk-go/rest"
)

// UsersGetter has a method to return a UserInterface.
// A group's client should implement this interface.
type UsersGetter interface {
	Users() UserInterface
}

// UserInterface has methods to work with User resources.
type UserInterface interface {
	Get(ctx context.Context, id string, opts metav1.GetOptions) (*v1.DetailUserResponse, error)
	Create(ctx context.Context, user *v1.CreateUserRequest, opts metav1.CreateOptions) (*v1.CreateUserResponse, error)
	Update(
		ctx context.Context,
		id string,
		user *v1.UpdateUserRequest,
		opts metav1.UpdateOptions,
	) (*v1.UpdateUserResponse, error)
	Delete(ctx context.Context, id string, opts metav1.DeleteOptions) error
	List(ctx context.Context, opts metav1.ListOptions) (*v1.UserList, error)
	Disable(ctx context.Context, id string) error
	Enable(ctx context.Context, id string) error
	UserExpansion
}

// users implements UserInterface.
type users struct {
	client rest.Interface
}

// newUsers returns a Users.
func newUsers(c *APIV1Client) *users {
	return &users{
		client: c.RESTClient(),
	}
}

// Get get user details
func (c *users) Get(ctx context.Context, id string, opts metav1.GetOptions) (result *v1.DetailUserResponse, err error) {
	result = &v1.DetailUserResponse{}
	err = c.client.Get().
		Resource("users").
		VersionedParams(opts).
		ID(id).
		Do(ctx).
		Into(result)

	return
}

// Create takes the representation of a user and creates it.
// Returns the server's representation of the user, and an error, if there is any.
func (c *users) Create(
	ctx context.Context,
	user *v1.CreateUserRequest,
	opts metav1.CreateOptions,
) (result *v1.CreateUserResponse, err error) {
	result = &v1.CreateUserResponse{}
	err = c.client.Post().
		Resource("users").
		VersionedParams(opts).
		Body(user).
		Do(ctx).
		Into(result)

	return
}

// Update takes the representation of a user and updates it.
// Returns the server's representation of the user, and an error, if there is any.
func (c *users) Update(
	ctx context.Context,
	id string,
	user *v1.UpdateUserRequest,
	opts metav1.UpdateOptions,
) (result *v1.UpdateUserResponse, err error) {
	result = &v1.UpdateUserResponse{}
	err = c.client.Put().
		Resource("users").
		VersionedParams(opts).
		ID(id).
		Body(user).
		Do(ctx).
		Into(result)

	return
}

// Delete delete a user
func (c *users) Delete(ctx context.Context, id string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Resource("users").
		ID(id).
		Body(&opts).
		Do(ctx).
		Error()
}

// List fetch users
func (c *users) List(ctx context.Context, opts metav1.ListOptions) (result *v1.UserList, err error) {
	result = &v1.UserList{}
	err = c.client.Get().
		Resource("users").
		VersionedParams(opts).
		Do(ctx).
		Into(result)

	return
}

// Disable disable user
func (c *users) Disable(ctx context.Context, id string) error {
	return c.client.Get().
		Resource("users").
		ID(id).
		SubResource("disable").
		Do(ctx).
		Error()
}

// Enable enable user
func (c *users) Enable(ctx context.Context, id string) error {
	return c.client.Get().
		Resource("users").
		ID(id).
		SubResource("enable").
		Do(ctx).
		Error()
}
