// Copied in and hacked up from the standard library to enable client
// timeouts.
//
// Those alterations are Copyright Heroku 2013.  All rights reserved.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// HTTP server.  See RFC 2616.

// TODO(rsc):
//	logging

package roundtripper

import (
	"bufio"
	"crypto/tls"
	"io"
	"net"
	"net/http"
)

// A conn represents the server side of an HTTP connection.
type conn struct {
	remoteAddr string               // network address of remote side
	server     *http.Server         // the Server on which the connection arrived
	rwc        net.Conn             // i/o connection
	lr         *io.LimitedReader    // io.LimitReader(rwc)
	buf        *bufio.ReadWriter    // buffered(lr,rwc), reading from bufio->limitReader->rwc
	hijacked   bool                 // connection has been hijacked by handler
	tlsState   *tls.ConnectionState // or nil when not using TLS
	body       []byte
}

// A response represents the server side of an HTTP response.
type response struct {
	conn          *conn
	req           *http.Request // request for this response
	chunking      bool          // using chunked transfer encoding for reply body
	wroteHeader   bool          // reply header has been written
	wroteContinue bool          // 100 Continue response was written
	header        http.Header   // reply header parameters
	written       int64         // number of bytes written in body
	contentLength int64         // explicitly-declared Content-Length; or -1
	status        int           // status code passed to WriteHeader
	needSniff     bool          // need to sniff to find Content-Type

	// close connection after this reply.  set on request and
	// updated after response from handler if there's a
	// "Connection: keep-alive" response header and a
	// Content-Length.
	closeAfterReply bool

	// requestBodyLimitHit is set by requestTooLarge when
	// maxBytesReader hits its max size. It is checked in
	// WriteHeader, to make sure we don't consume the the
	// remaining request body to try to advance to the next HTTP
	// request. Instead, when this is set, we stop doing
	// subsequent requests on this connection and stop reading
	// input from it.
	requestBodyLimitHit bool
}
