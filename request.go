// Copied in and hacked up from the standard library to enable client
// timeouts.
//
// Those alterations are Copyright Heroku 2013.  All rights reserved.
//
// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// HTTP Request reading and parsing.
package roundtripper

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type badStringError struct {
	what string
	str  string
}

func (e *badStringError) Error() string { return fmt.Sprintf("%s %q", e.what, e.str) }

// Return value if nonempty, def otherwise.
func valueOrDefault(value, def string) string {
	if value != "" {
		return value
	}
	return def
}

// Headers that Request.Write handles itself and should be skipped.
var reqWriteExcludeHeader = map[string]bool{
	"Host":              true, // not in Header map anyway
	"User-Agent":        true,
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Trailer":           true,
}

const defaultUserAgent = "Go http package"

// extraHeaders may be nil
func write(req *http.Request, w io.Writer, usingProxy bool, extraHeaders http.Header) error {
	host := req.Host
	if host == "" {
		if req.URL == nil {
			return errors.New("http: Request.Write on Request with no Host or URL set")
		}
		host = req.URL.Host
	}

	ruri := req.URL.RequestURI()
	if usingProxy && req.URL.Scheme != "" && req.URL.Opaque == "" {
		ruri = req.URL.Scheme + "://" + host + ruri
	} else if req.Method == "CONNECT" && req.URL.Path == "" {
		// CONNECT requests normally give just the host and port, not a full URL.
		ruri = host
	}
	// TODO(bradfitz): escape at least newlines in ruri?

	bw := bufio.NewWriter(w)
	fmt.Fprintf(bw, "%s %s HTTP/1.1\r\n", valueOrDefault(req.Method, "GET"), ruri)

	// Header lines
	fmt.Fprintf(bw, "Host: %s\r\n", host)

	// Use the defaultUserAgent unless the Header contains one, which
	// may be blank to not send the header.
	userAgent := defaultUserAgent
	if req.Header != nil {
		if ua := req.Header["User-Agent"]; len(ua) > 0 {
			userAgent = ua[0]
		}
	}
	if userAgent != "" {
		fmt.Fprintf(bw, "User-Agent: %s\r\n", userAgent)
	}

	// Process Body,ContentLength,Close,Trailer
	tw, err := newTransferWriter(req)
	if err != nil {
		return err
	}
	err = tw.WriteHeader(bw)
	if err != nil {
		return err
	}

	// TODO: split long values?  (If so, should share code with Conn.Write)
	err = req.Header.WriteSubset(bw, reqWriteExcludeHeader)
	if err != nil {
		return err
	}

	if extraHeaders != nil {
		err = extraHeaders.Write(bw)
		if err != nil {
			return err
		}
	}

	io.WriteString(bw, "\r\n")

	// Write body and trailer
	err = tw.WriteBody(bw)
	if err != nil {
		return err
	}

	return bw.Flush()
}
