// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Note: the example only works with the code within the same release/branch.
package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/AlekSi/pointer"

	metav1 "github.com/coding-hui/common/meta/v1"
	"github.com/coding-hui/common/util/homedir"
	v1 "github.com/coding-hui/iam/pkg/api/apiserver/v1"

	"github.com/coding-hui/wecoding-sdk-go/services/iam"
	"github.com/coding-hui/wecoding-sdk-go/tools/clientcmd"
)

func main() {
	var iamconfig *string
	if home := homedir.HomeDir(); home != "" {
		iamconfig = flag.String(
			"iamconfig",
			filepath.Join(home, ".iam", "config"),
			"(optional) absolute path to the iamconfig file",
		)
	} else {
		iamconfig = flag.String("iamconfig", "", "absolute path to the iamconfig file")
	}
	flag.Parse()

	// use the current context in iamconfig
	config, err := clientcmd.BuildConfigFromFlags("", *iamconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the iamclient
	iamclient, err := iam.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	iamclient.APIV1().RESTClient().Get()

	usersClient := iamclient.APIV1().Users()

	user := &v1.CreateUserRequest{
		Name:     "sdk",
		Alias:    "sdkexample",
		Password: "Sdk@2023",
		Email:    "user@qq.com",
		Phone:    "1235125xxxx",
	}

	// Create user
	fmt.Println("Creating user...")
	createReqp, err := usersClient.Create(context.TODO(), user, metav1.CreateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created user %q.\n", createReqp.GetObjectMeta().GetName())
	defer func() {
		// Delete secret
		fmt.Println("Deleting user...")
		if err := usersClient.Delete(context.TODO(), createReqp.InstanceID, metav1.DeleteOptions{}); err != nil {
			fmt.Printf("Delete user failed: %s\n", err.Error())
			return
		}
		fmt.Println("Deleted user.")
	}()

	// Get user
	prompt()
	fmt.Println("Geting user...")
	createdUser, err := usersClient.Get(context.TODO(), createReqp.InstanceID, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Get user %q.\n", createdUser.GetObjectMeta().GetName())

	// Update user
	prompt()
	fmt.Println("Updating user...")
	updateReq := &v1.UpdateUserRequest{
		Alias: "sdkexample_update",
		Email: "user_update@qq.com",
		Phone: "1812885xxxx",
	}
	updateResp, err := usersClient.Update(context.TODO(), createReqp.InstanceID, updateReq, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Created user %q.\n", updateResp.GetObjectMeta().GetName())

	prompt()
	fmt.Println("Disable user...")
	err = usersClient.Disable(context.TODO(), createdUser.InstanceID)
	if err != nil {
		panic(err.Error())
	}

	prompt()
	fmt.Println("Enable user...")
	err = usersClient.Enable(context.TODO(), createdUser.InstanceID)
	if err != nil {
		panic(err.Error())
	}

	// List users
	prompt()
	fmt.Println("List users...")
	listOpts := metav1.ListOptions{
		Offset: pointer.ToInt64(5),
		Limit:  pointer.ToInt64(5),
	}
	users, err := usersClient.List(context.TODO(), listOpts)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Fetch [%d] users, total [%d].\n", len(users.Items), users.TotalCount)
}

func prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}
