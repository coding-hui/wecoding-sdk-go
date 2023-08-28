// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Note: the example only works with the code within the same release/branch.
package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"

	"github.com/coding-hui/common/util/homedir"

	v1 "github.com/coding-hui/iam/pkg/api/authzserver/v1"

	"github.com/coding-hui/wecoding-sdk-go/tools/clientcmd"
	"github.com/coding-hui/wecoding-sdk-go/wecoding"
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

	// create the clientset
	clientset, err := wecoding.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	request := &v1.Request{
		Resource: "project:sdk",
		Action:   "delete",
		Subject:  "users:sdk-user",
	}

	// Authorize the request
	fmt.Println("Authorize request...")
	ret, err := clientset.Iam().AuthzV1().Authz().Authorize(context.TODO(), request)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Authorize response: %s.\n", ret.ToString())
}
