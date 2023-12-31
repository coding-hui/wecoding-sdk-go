// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

//go:build go1.4 && !go1.6
// +build go1.4,!go1.6

package gorequest

import (
	"net/http"
)

// does a shallow clone of the transport
func (s *SuperAgent) safeModifyTransport() {
	if !s.isClone {
		return
	}
	oldTransport := s.Transport
	s.Transport = &http.Transport{
		Proxy:                 oldTransport.Proxy,
		Dial:                  oldTransport.Dial,
		TLSClientConfig:       oldTransport.TLSClientConfig,
		TLSHandshakeTimeout:   oldTransport.TLSHandshakeTimeout,
		DisableKeepAlives:     oldTransport.DisableKeepAlives,
		DisableCompression:    oldTransport.DisableCompression,
		MaxIdleConnsPerHost:   oldTransport.MaxIdleConnsPerHost,
		ResponseHeaderTimeout: oldTransport.ResponseHeaderTimeout,
		// new in go1.4
		DialTLS: oldTransport.DialTLS,
	}
}
