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

	authClient := iamclient.APIV1().Authentication()

	// Login
	fmt.Println("Login...")
	authReqp, err := authClient.Authenticate(context.TODO(), v1.AuthenticateRequest{
		Username: "ADMIN",
		Password: "WECODING",
	})
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Login success %q.\n", authReqp.User.Name)

	// Get user-info
	prompt()
	fmt.Println("Geting user-info...")
	userInfo, err := authClient.UserInfo(context.TODO(), authReqp.AccessToken)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("Get user %q.\n", userInfo.GetName())

	// Refresh token
	prompt()
	fmt.Println("Refresh token...")
	refreshResp, err := authClient.RefreshToken(context.TODO(), authReqp.RefreshToken)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("AccessToken %q.\n", refreshResp.AccessToken)
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
