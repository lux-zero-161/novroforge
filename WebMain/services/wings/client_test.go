// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

const testServerID = "7f6f323c-6e25-4b79-a3d8-3d506965851d"

func TestNewClientRejectsInsecureOrMalformedURLs(t *testing.T) {
	t.Parallel()

	for _, nodeURL := range []string{
		"http://node.example.com",
		"ftp://localhost:8080",
		"https://user:password@localhost",
		"https://localhost/path",
		"https://localhost?token=secret",
	} {
		t.Run(nodeURL, func(t *testing.T) {
			t.Parallel()
			_, err := NewClient(nodeURL, "node-token")
			require.Error(t, err)
		})
	}
}

func TestNewClientRejectsUnsafeTokens(t *testing.T) {
	t.Parallel()

	for _, token := range []string{"", " spaced", "line\nbreak", "bad\x7f", strings.Repeat("x", maxTokenLength+1)} {
		_, err := NewClient("https://node.example.com", token)
		require.Error(t, err)
	}
}

func TestGetServerUsesBearerTokenAndDecodesDetails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer node-token" {
			t.Errorf("authorization header = %q, want bearer token", got)
		}
		if got := r.URL.Path; got != "/api/servers/"+testServerID {
			t.Errorf("request path = %q, want server path", got)
		}
		if err := json.NewEncoder(w).Encode(ServerDetails{
			State:       "running",
			IsSuspended: false,
			Utilization: Utilization{MemoryBytes: 512, CPUAbsolute: 7.5},
		}); err != nil {
			t.Errorf("encode test response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	details, err := client.GetServer(context.Background(), testServerID)
	require.NoError(t, err)
	require.Equal(t, "running", details.State)
	require.Equal(t, uint64(512), details.Utilization.MemoryBytes)
	require.Equal(t, 7.5, details.Utilization.CPUAbsolute)
}

func TestGetSystemUsesBearerToken(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/api/system" {
			t.Errorf("request path = %q, want /api/system", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer node-token" {
			t.Errorf("authorization header = %q, want bearer token", got)
		}
		if _, err := w.Write([]byte(`{"architecture":"amd64","cpu_count":8,"kernel_version":"test","os":"linux","version":"1.0"}`)); err != nil {
			t.Errorf("write test response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	info, err := client.GetSystem(context.Background())
	require.NoError(t, err)
	require.Equal(t, "linux", info.OS)
	require.Equal(t, "1.0", info.Version)
	require.Equal(t, 8, info.CPUCount)
}

func TestPowerRejectsUnsupportedActions(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://127.0.0.1:8080", "node-token")
	require.NoError(t, err)
	require.Error(t, client.Power(context.Background(), testServerID, "execute"))
	require.Error(t, client.Power(context.Background(), "not-a-uuid", "start"))
}

func TestPowerSendsExpectedAction(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/power" {
			t.Errorf("request path = %q, want power path", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer node-token" {
			t.Errorf("authorization header = %q, want bearer token", got)
		}
		var body struct {
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if body.Action != "restart" {
			t.Errorf("action = %q, want restart", body.Action)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.Power(context.Background(), testServerID, "restart"))
}

func TestLogsValidatesLineLimit(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	_, err = client.Logs(context.Background(), testServerID, 101)
	require.Error(t, err)
}

func TestLogsRequestsAndDecodesRecentLines(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %q, want GET", r.Method)
		}
		if r.URL.Path != "/api/servers/"+testServerID+"/logs" {
			t.Errorf("request path = %q, want server logs path", r.URL.Path)
		}
		if got := r.URL.Query().Get("size"); got != "100" {
			t.Errorf("log size = %q, want 100", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer node-token" {
			t.Errorf("authorization header = %q, want bearer token", got)
		}
		if _, err := w.Write([]byte(`{"data":["server started","ready"]}`)); err != nil {
			t.Errorf("write test response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	lines, err := client.Logs(context.Background(), testServerID, 100)
	require.NoError(t, err)
	require.Equal(t, []string{"server started", "ready"}, lines)
}

func TestListDirectoryRequestsAndDecodesEntries(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/list-directory" {
			t.Errorf("request path = %q, want server file listing path", got)
		}
		if got := r.URL.Query().Get("directory"); got != "/world one" {
			t.Errorf("directory query = %q, want %q", got, "/world one")
		}
		if _, err := w.Write([]byte(`[{"name":"world one","directory":true,"file":false,"symlink":false,"size":0}]`)); err != nil {
			t.Errorf("write test response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	entries, err := client.ListDirectory(context.Background(), testServerID, "world one")
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.Equal(t, "world one", entries[0].Name)
	require.True(t, entries[0].Directory)
}

func TestListDirectoryRejectsUnsafePaths(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	for _, directory := range []string{"../outside", "world/../../outside", `world\outside`, strings.Repeat("a", maxDirectoryPath+1)} {
		_, err := client.ListDirectory(context.Background(), testServerID, directory)
		require.ErrorIs(t, err, ErrInvalidDirectory)
	}
}

func TestFileContentsReturnsTextPreview(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/contents" {
			t.Errorf("request path = %q, want server file contents path", got)
		}
		if got := r.URL.Query().Get("file"); got != "/server.properties" {
			t.Errorf("file query = %q, want %q", got, "/server.properties")
		}
		if got := r.Header.Get("Authorization"); got != "******" {
			t.Errorf("authorization header = %q, want bearer token", got)
		}
		w.Header().Set("X-Mime-Type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("motd=hello"))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	content, err := client.FileContents(context.Background(), testServerID, "server.properties")
	require.NoError(t, err)
	require.Equal(t, "motd=hello", content)
}

func TestFileContentsRejectsInvalidOrUnsupportedFiles(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name        string
		filePath    string
		contentType string
		content     []byte
		wantErr     error
	}{
		{name: "path traversal", filePath: "../outside", wantErr: ErrInvalidFilePath},
		{name: "directory path", filePath: "configs/", wantErr: ErrInvalidFilePath},
		{name: "binary MIME", filePath: "image.bin", contentType: "application/octet-stream", content: []byte{0, 1}, wantErr: ErrUnsupportedFileType},
		{name: "invalid UTF-8", filePath: "text.txt", contentType: "text/plain", content: []byte{0xff}, wantErr: ErrUnsupportedFileType},
		{name: "oversized preview", filePath: "large.txt", contentType: "text/plain", content: make([]byte, maxFilePreviewSize+1), wantErr: ErrFileTooLarge},
	} {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("X-Mime-Type", test.contentType)
				_, _ = w.Write(test.content)
			}))
			defer server.Close()

			client, err := NewClient(server.URL, "node-token")
			require.NoError(t, err)
			_, err = client.FileContents(context.Background(), testServerID, test.filePath)
			require.ErrorIs(t, err, test.wantErr)
		})
	}
}

func TestNormalizeDirectoryPath(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		input string
		want  string
	}{
		{input: "", want: "/"},
		{input: "/", want: "/"},
		{input: "worlds/survival", want: "/worlds/survival"},
		{input: "/worlds//survival/", want: "/worlds/survival"},
	} {
		got, err := NormalizeDirectoryPath(test.input)
		require.NoError(t, err)
		require.Equal(t, test.want, got)
	}
}

func TestCommandsRejectsEmptyCommands(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	require.Error(t, client.Commands(context.Background(), testServerID, []string{" "}))
	require.Error(t, client.Commands(context.Background(), testServerID, make([]string, maxCommands+1)))
}

func TestClientDoesNotFollowRedirects(t *testing.T) {
	t.Parallel()

	var redirected atomic.Bool
	target := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		redirected.Store(true)
	}))
	defer target.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	_, err = client.GetServer(context.Background(), testServerID)
	require.Error(t, err)
	require.False(t, redirected.Load())
}
