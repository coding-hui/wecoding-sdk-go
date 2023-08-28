// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build go1.3
// +build go1.3

package gorequest

import (
	"net/http"
	"time"
)

// we don't want to mess up other clones when we modify the client..
// so unfortantely we need to create a new client
func (s *SuperAgent) safeModifyHttpClient() {
	if !s.isClone {
		return
	}
	oldClient := s.Client
	s.Client = &http.Client{}
	s.Client.Jar = oldClient.Jar
	s.Client.Transport = oldClient.Transport
	s.Client.Timeout = oldClient.Timeout
	s.Client.CheckRedirect = oldClient.CheckRedirect
}

func (s *SuperAgent) Timeout(timeout time.Duration) *SuperAgent {
	s.safeModifyHttpClient()
	s.Client.Timeout = timeout
	return s
}
