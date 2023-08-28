// Copyright (c) 2023 coding-hui. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gorequest

type Logger interface {
	SetPrefix(string)
	Printf(format string, v ...interface{})
	Println(v ...interface{})
}
