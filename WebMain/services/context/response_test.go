// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: MIT

package context

import (
	"bufio"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

type hijackableResponseWriter struct {
	*httptest.ResponseRecorder
	connection net.Conn
	buffered   *bufio.ReadWriter
}

func (w *hijackableResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return w.connection, w.buffered, nil
}

func TestResponseHijackRecordsSwitchingProtocols(t *testing.T) {
	serverConnection, peerConnection := net.Pipe()
	defer serverConnection.Close()
	defer peerConnection.Close()

	writer := &hijackableResponseWriter{
		ResponseRecorder: httptest.NewRecorder(),
		connection:      serverConnection,
		buffered:        bufio.NewReadWriter(bufio.NewReader(serverConnection), bufio.NewWriter(serverConnection)),
	}
	response := WrapResponseWriter(writer)

	connection, buffered, err := response.Hijack()
	if err != nil {
		t.Fatalf("hijack response: %v", err)
	}
	if connection != serverConnection || buffered != writer.buffered {
		t.Fatal("Hijack did not return the underlying connection and buffers")
	}
	if response.Status() != http.StatusSwitchingProtocols || response.WrittenStatus() != http.StatusSwitchingProtocols {
		t.Fatalf("response status = %d/%d, want 101", response.Status(), response.WrittenStatus())
	}
	if response.Unwrap() != writer {
		t.Fatal("Unwrap did not return the underlying response writer")
	}
}
