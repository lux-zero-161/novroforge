// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package migration

import "net/http"

type UserAgentTransport struct {
	*http.Transport
}

func (t *UserAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "NovroForge-migration-http-client/1.1")
	return t.Transport.RoundTrip(req)
}
