// Copied in and hacked up from the standard library to enable client
// timeouts.
//
// Those alterations are Copyright Heroku 2013.  All rights reserved.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// HTTP client. See RFC 2616.
// 
// This is the high-level Client interface.
// The low-level implementation is in transport.go.
package roundtripper

import (
	"strings"
)

// Given a string of the form "host", "host:port", or "[ipv6::address]:port",
// return true if the string includes a port.
func hasPort(s string) bool { return strings.LastIndex(s, ":") > strings.LastIndex(s, "]") }
