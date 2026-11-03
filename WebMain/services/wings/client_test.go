// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gorilla/websocket"
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
	require.Equal(t, "amd64", info.Architecture)
	require.Equal(t, "test", info.KernelVersion)
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

func TestPowerSendsKillAction(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/servers/"+testServerID+"/power" {
			t.Errorf("power request = %s %s", r.Method, r.URL.Path)
		}
		var body struct {
			Action string `json:"action"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		if body.Action != "kill" {
			t.Errorf("action = %q, want kill", body.Action)
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.Power(context.Background(), testServerID, "kill"))
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
		if got := r.Header.Get("Authorization"); got != "Bearer node-token" {
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

func TestFileDownloadStreamsFile(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/contents" {
			t.Errorf("request path = %q, want server file contents path", got)
		}
		if got := r.URL.Query().Get("file"); got != "/logs/server.log" {
			t.Errorf("file query = %q, want %q", got, "/logs/server.log")
		}
		if got := r.Header.Get("Authorization"); got != "Bearer node-token" {
			t.Errorf("authorization header = %q, want bearer token", got)
		}
		time.Sleep(30 * time.Millisecond)
		_, _ = w.Write([]byte("server log contents"))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	client.httpClient.Timeout = time.Millisecond
	response, err := client.FileDownload(context.Background(), testServerID, "logs/server.log")
	require.NoError(t, err)
	defer response.Body.Close()

	content, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, "server log contents", string(content))
}

func TestFileDownloadRejectsInvalidPaths(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	for _, filePath := range []string{"../outside", "configs/", `configs\file.txt`, "/"} {
		_, err := client.FileDownload(context.Background(), testServerID, filePath)
		require.ErrorIs(t, err, ErrInvalidFilePath)
	}
}

func TestWriteFileSendsBoundedRawContent(t *testing.T) {
	t.Parallel()

	const expectedContent = "server-name=novroforge\nmax-players=20\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPost {
			t.Errorf("request method = %q, want POST", got)
		}
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/write" {
			t.Errorf("request path = %q, want file write path", got)
		}
		if got := r.URL.Query().Get("file"); got != "/server.properties" {
			t.Errorf("file query = %q, want %q", got, "/server.properties")
		}
		if got := r.Header.Get("Content-Type"); got != "application/octet-stream" {
			t.Errorf("content type = %q, want application/octet-stream", got)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read request body: %v", err)
		}
		if string(body) != expectedContent {
			t.Errorf("request body = %q, want exact file content", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.WriteFile(context.Background(), testServerID, "server.properties", expectedContent))
}

func TestWriteFileRejectsInvalidContentAndPaths(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	require.ErrorIs(t, client.WriteFile(context.Background(), testServerID, "../outside", "content"), ErrInvalidFilePath)
	require.ErrorIs(t, client.WriteFile(context.Background(), testServerID, "file.txt", string([]byte{0xff})), ErrUnsupportedFileType)
	require.ErrorIs(t, client.WriteFile(context.Background(), testServerID, "large.txt", strings.Repeat("a", maxFilePreviewSize+1)), ErrFileTooLarge)
}

func TestCreateDirectorySendsValidatedRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/create-directory" {
			t.Errorf("request path = %q, want create-directory path", got)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer node-token" {
			t.Errorf("authorization header = %q, want bearer token", got)
		}
		var body struct {
			Name string `json:"name"`
			Path string `json:"path"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode create-directory body: %v", err)
		}
		if body.Name != "new folder" || body.Path != "/worlds/survival" {
			t.Errorf("create-directory body = %#v, want normalized path and name", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.CreateDirectory(context.Background(), testServerID, "new folder", "worlds//survival/"))
}

func TestCreateDirectoryRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	for _, name := range []string{"", ".", "..", "has/slash", `has\slash`, " trailing", "trailing ", strings.Repeat("a", maxFileNameLength+1)} {
		require.ErrorIs(t, client.CreateDirectory(context.Background(), testServerID, name, "/"), ErrInvalidDirectoryName)
	}
	require.ErrorIs(t, client.CreateDirectory(context.Background(), testServerID, "valid", "../outside"), ErrInvalidDirectory)
}

func TestRenameFileSendsValidatedRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPut {
			t.Errorf("request method = %q, want PUT", got)
		}
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/rename" {
			t.Errorf("request path = %q, want file rename path", got)
		}
		var body struct {
			Root  string `json:"root"`
			Files []struct {
				From string `json:"from"`
				To   string `json:"to"`
			} `json:"files"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode rename body: %v", err)
		}
		if body.Root != "/worlds/survival" || len(body.Files) != 1 || body.Files[0].From != "old.txt" || body.Files[0].To != "new.txt" {
			t.Errorf("rename body = %#v, want normalized directory and single rename", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.RenameFile(context.Background(), testServerID, "worlds//survival/", "old.txt", "new.txt"))
}

func TestRenameFileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	require.ErrorIs(t, client.RenameFile(context.Background(), testServerID, "../outside", "old.txt", "new.txt"), ErrInvalidDirectory)
	require.ErrorIs(t, client.RenameFile(context.Background(), testServerID, "/", "../outside", "new.txt"), ErrInvalidDirectoryName)
	require.ErrorIs(t, client.RenameFile(context.Background(), testServerID, "/", "old.txt", "old.txt"), ErrInvalidDirectoryName)
}

func TestDeleteFileSendsOneValidatedFile(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPost {
			t.Errorf("request method = %q, want POST", got)
		}
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/delete" {
			t.Errorf("request path = %q, want file delete path", got)
		}
		var body struct {
			Root  string   `json:"root"`
			Files []string `json:"files"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode delete body: %v", err)
		}
		if body.Root != "/worlds/survival" || len(body.Files) != 1 || body.Files[0] != "old.txt" {
			t.Errorf("delete body = %#v, want normalized directory and one file", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.DeleteFile(context.Background(), testServerID, "worlds//survival/", "old.txt"))
}

func TestDeleteFileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	require.Error(t, client.DeleteFile(context.Background(), "invalid", "/", "old.txt"))
	require.ErrorIs(t, client.DeleteFile(context.Background(), testServerID, "../outside", "old.txt"), ErrInvalidDirectory)
	for _, name := range []string{"", ".", "..", "has/slash", `has\slash`, " trailing", "trailing ", strings.Repeat("a", maxFileNameLength+1)} {
		require.ErrorIs(t, client.DeleteFile(context.Background(), testServerID, "/", name), ErrInvalidDirectoryName)
	}
}

func TestCopyFileSendsValidatedRequest(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPost {
			t.Errorf("request method = %q, want POST", got)
		}
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/copy" {
			t.Errorf("request path = %q, want file copy path", got)
		}
		var body struct {
			Location string `json:"location"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode copy body: %v", err)
		}
		if body.Location != "/worlds/survival/world.tar.gz" {
			t.Errorf("copy location = %q, want normalized file path", body.Location)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.CopyFile(context.Background(), testServerID, "worlds//survival/world.tar.gz"))
}

func TestCopyFileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	require.Error(t, client.CopyFile(context.Background(), "invalid", "file.txt"))
	require.ErrorIs(t, client.CopyFile(context.Background(), testServerID, "../outside"), ErrInvalidFilePath)
	require.ErrorIs(t, client.CopyFile(context.Background(), testServerID, "/"), ErrInvalidFilePath)
}

func TestCompressFileSendsOneValidatedEntry(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPost {
			t.Errorf("request method = %q, want POST", got)
		}
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/compress" {
			t.Errorf("request path = %q, want file compress path", got)
		}
		var body struct {
			Root  string   `json:"root"`
			Files []string `json:"files"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode compress body: %v", err)
		}
		if body.Root != "/worlds/survival" || len(body.Files) != 1 || body.Files[0] != "server-config" {
			t.Errorf("compress body = %#v, want normalized directory and one selected entry", body)
		}
		time.Sleep(20 * time.Millisecond)
		if err := json.NewEncoder(w).Encode(FileEntry{Name: "archive-2026.tar.gz", File: true, Size: 128}); err != nil {
			t.Errorf("encode archive response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	client.httpClient.Timeout = time.Millisecond
	archive, err := client.CompressFile(context.Background(), testServerID, "worlds//survival/", "server-config")
	require.NoError(t, err)
	require.Equal(t, "archive-2026.tar.gz", archive.Name)
	require.Equal(t, int64(128), archive.Size)
}

func TestCompressFileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	_, err = client.CompressFile(context.Background(), "invalid", "/", "server.properties")
	require.Error(t, err)
	_, err = client.CompressFile(context.Background(), testServerID, "../outside", "server.properties")
	require.ErrorIs(t, err, ErrInvalidDirectory)
	_, err = client.CompressFile(context.Background(), testServerID, "/", "../outside")
	require.ErrorIs(t, err, ErrInvalidDirectoryName)
}

func TestBackupOperationsUseLocalWingsEndpoints(t *testing.T) {
	t.Parallel()

	const backupID = "9a7b3db9-0630-4b61-8cec-cda7e0ca4453"
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		switch requests {
		case 1:
			if r.Method != http.MethodPost || r.URL.Path != "/api/servers/"+testServerID+"/backup" {
				t.Errorf("create backup request = %s %s", r.Method, r.URL.Path)
			}
			var body struct {
				Adapter string `json:"adapter"`
				UUID    string `json:"uuid"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode create backup request: %v", err)
			}
			if body.Adapter != "wings" || body.UUID != backupID {
				t.Errorf("create backup body = %#v", body)
			}
			w.WriteHeader(http.StatusAccepted)
		case 2:
			if r.Method != http.MethodPost || r.URL.Path != "/api/servers/"+testServerID+"/backup/"+backupID+"/restore" {
				t.Errorf("restore backup request = %s %s", r.Method, r.URL.Path)
			}
			var body struct {
				Adapter           string `json:"adapter"`
				TruncateDirectory bool   `json:"truncate_directory"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Errorf("decode restore backup request: %v", err)
			}
			if body.Adapter != "wings" || body.TruncateDirectory {
				t.Errorf("restore backup body = %#v, want local adapter without truncation", body)
			}
			w.WriteHeader(http.StatusAccepted)
		case 3:
			if r.Method != http.MethodDelete || r.URL.Path != "/api/servers/"+testServerID+"/backup/"+backupID {
				t.Errorf("delete backup request = %s %s", r.Method, r.URL.Path)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected backup request %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.CreateBackup(context.Background(), testServerID, backupID))
	require.NoError(t, client.RestoreBackup(context.Background(), testServerID, backupID))
	require.NoError(t, client.DeleteBackup(context.Background(), testServerID, backupID))
	require.Equal(t, 3, requests)
}

func TestBackupOperationsRejectInvalidUUIDs(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	require.Error(t, client.CreateBackup(context.Background(), "invalid", "invalid"))
	require.Error(t, client.RestoreBackup(context.Background(), testServerID, "invalid"))
	require.Error(t, client.DeleteBackup(context.Background(), "invalid", testServerID))
}

func TestDownloadBackupUsesScopedSingleUseTokenWithoutDaemonAuthorization(t *testing.T) {
	t.Parallel()

	const (
		userID   = "f1b3d5a7-1234-4567-89ab-cdef01234567"
		backupID = "9a7b3db9-0630-4b61-8cec-cda7e0ca4453"
		content  = "backup archive"
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/download/backup" {
			t.Errorf("download request = %s %s, want GET /download/backup", r.Method, r.URL.Path)
		}
		if authorization := r.Header.Get("Authorization"); authorization != "" {
			t.Errorf("download authorization = %q, want no daemon credential", authorization)
		}
		claims := decodeTestTokenClaims(t, r.URL.Query().Get("token"), "node-token")
		if claims.Scope != "backup-download" || claims.ServerUUID != testServerID || claims.UserUUID != userID || claims.BackupUUID != backupID || claims.UniqueID == "" || claims.JWTID != claims.UniqueID {
			t.Errorf("backup download claims = %#v, want scoped single-use server/user/backup token", claims)
		}
		if claims.ExpirationTime <= claims.IssuedAt || claims.ExpirationTime-claims.IssuedAt > 60 {
			t.Errorf("backup download token lifetime = %d seconds, want at most 60", claims.ExpirationTime-claims.IssuedAt)
		}
		w.Header().Set("Content-Length", strconv.Itoa(len(content)))
		_, _ = w.Write([]byte(content))
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	response, err := client.DownloadBackup(context.Background(), testServerID, userID, backupID)
	require.NoError(t, err)
	defer response.Body.Close()
	got, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.Equal(t, content, string(got))
	require.Equal(t, int64(len(content)), response.ContentLength)
}

func TestDownloadBackupRejectsInvalidUUIDs(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	_, err = client.DownloadBackup(context.Background(), "invalid", testServerID, testServerID)
	require.Error(t, err)
	_, err = client.DownloadBackup(context.Background(), testServerID, "invalid", testServerID)
	require.Error(t, err)
	_, err = client.DownloadBackup(context.Background(), testServerID, testServerID, "invalid")
	require.Error(t, err)
}

func TestDownloadBackupRedactsSignedTokenFromNetworkErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	_, err = client.DownloadBackup(context.Background(), testServerID, "f1b3d5a7-1234-4567-89ab-cdef01234567", "9a7b3db9-0630-4b61-8cec-cda7e0ca4453")
	require.Error(t, err)
	require.NotContains(t, err.Error(), "token=")
	require.NotContains(t, err.Error(), "backup-download")
}

func TestOpenConsoleWebSocketUsesShortLivedReadOnlyToken(t *testing.T) {
	t.Parallel()

	const (
		userID = "f1b3d5a7-1234-4567-89ab-cdef01234567"
		origin = "https://panel.example.test"
	)
	tokenResult := make(chan string, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/servers/"+testServerID+"/ws" || r.Header.Get("Origin") != origin || r.Header.Get("Authorization") != "" {
			t.Errorf("websocket request path/origin/authorization = %q/%q/%q", r.URL.Path, r.Header.Get("Origin"), r.Header.Get("Authorization"))
			http.Error(w, "unexpected websocket request", http.StatusBadRequest)
			return
		}
		connection, err := websocket.Upgrader{}.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade websocket: %v", err)
			return
		}
		defer connection.Close()

		var message struct {
			Event string   `json:"event"`
			Args  []string `json:"args"`
		}
		if err := connection.ReadJSON(&message); err != nil {
			t.Errorf("read websocket auth message: %v", err)
			return
		}
		if message.Event != "auth" || len(message.Args) != 1 {
			t.Errorf("websocket auth message = %#v", message)
			return
		}
		tokenResult <- message.Args[0]
		if err := connection.WriteJSON(map[string]string{"event": "auth success"}); err != nil {
			t.Errorf("write websocket auth response: %v", err)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	connection, err := client.OpenConsoleWebSocket(context.Background(), testServerID, userID, origin)
	require.NoError(t, err)
	require.NoError(t, connection.Close())
	claims := decodeTestTokenClaims(t, <-tokenResult, "node-token")
	require.Equal(t, "websocket", claims.Scope)
	require.Equal(t, testServerID, claims.ServerUUID)
	require.Equal(t, userID, claims.UserUUID)
	require.NotEmpty(t, claims.JWTID)
	require.Equal(t, []string{"websocket.connect"}, claims.Permissions)
	require.Greater(t, claims.IssuedAt, time.Now().Add(-time.Second).Unix())
	require.LessOrEqual(t, claims.ExpirationTime-claims.IssuedAt, int64(600))
}

func TestDeleteBackupTreatsMissingArchiveAsDeleted(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.DeleteBackup(context.Background(), testServerID, "9a7b3db9-0630-4b61-8cec-cda7e0ca4453"))
}

func decodeTestTokenClaims(t *testing.T, token, secret string) struct {
	IssuedAt       int64    `json:"iat"`
	ExpirationTime int64    `json:"exp"`
	JWTID          string   `json:"jti"`
	Scope          string   `json:"scope"`
	ServerUUID     string   `json:"server_uuid"`
	UserUUID       string   `json:"user_uuid"`
	BackupUUID     string   `json:"backup_uuid"`
	UniqueID       string   `json:"unique_id"`
	Permissions    []string `json:"permissions"`
} {
	t.Helper()
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	require.NoError(t, err)
	mac := hmac.New(sha256.New, []byte(secret))
	_, err = mac.Write([]byte(parts[0] + "." + parts[1]))
	require.NoError(t, err)
	require.True(t, hmac.Equal(signature, mac.Sum(nil)))
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)
	var claims struct {
		IssuedAt       int64    `json:"iat"`
		ExpirationTime int64    `json:"exp"`
		JWTID          string   `json:"jti"`
		Scope          string   `json:"scope"`
		ServerUUID     string   `json:"server_uuid"`
		UserUUID       string   `json:"user_uuid"`
		BackupUUID     string   `json:"backup_uuid"`
		UniqueID       string   `json:"unique_id"`
		Permissions    []string `json:"permissions"`
	}
	require.NoError(t, json.Unmarshal(payload, &claims))
	return claims
}

func TestDecompressFileSendsOneValidatedArchive(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPost {
			t.Errorf("request method = %q, want POST", got)
		}
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/decompress" {
			t.Errorf("request path = %q, want file decompress path", got)
		}
		var body struct {
			Root string `json:"root"`
			File string `json:"file"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode decompress body: %v", err)
		}
		if body.Root != "/worlds/survival" || body.File != "archive.tar.gz" {
			t.Errorf("decompress body = %#v, want normalized directory and archive", body)
		}
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	client.httpClient.Timeout = time.Millisecond
	require.NoError(t, client.DecompressFile(context.Background(), testServerID, "worlds//survival/", "archive.tar.gz"))
}

func TestDecompressFileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	require.Error(t, client.DecompressFile(context.Background(), "invalid", "/", "archive.zip"))
	require.ErrorIs(t, client.DecompressFile(context.Background(), testServerID, "../outside", "archive.zip"), ErrInvalidDirectory)
	require.ErrorIs(t, client.DecompressFile(context.Background(), testServerID, "/", "../outside"), ErrInvalidDirectoryName)
}

func TestChmodFileSendsOneValidatedRegularFile(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Method; got != http.MethodPost {
			t.Errorf("request method = %q, want POST", got)
		}
		if got := r.URL.Path; got != "/api/servers/"+testServerID+"/files/chmod" {
			t.Errorf("request path = %q, want file chmod path", got)
		}
		var body struct {
			Root  string `json:"root"`
			Files []struct {
				File string `json:"file"`
				Mode string `json:"mode"`
			} `json:"files"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode chmod body: %v", err)
		}
		if body.Root != "/worlds/survival" || len(body.Files) != 1 || body.Files[0].File != "start.sh" || body.Files[0].Mode != "755" {
			t.Errorf("chmod body = %#v, want normalized directory, one file, and canonical mode", body)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.ChmodFile(context.Background(), testServerID, "worlds//survival/", "start.sh", "0755"))
}

func TestChmodFileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	require.Error(t, client.ChmodFile(context.Background(), "invalid", "/", "start.sh", "0755"))
	require.ErrorIs(t, client.ChmodFile(context.Background(), testServerID, "../outside", "start.sh", "0755"), ErrInvalidDirectory)
	require.ErrorIs(t, client.ChmodFile(context.Background(), testServerID, "/", "../outside", "0755"), ErrInvalidDirectoryName)
	for _, mode := range []string{"", "7", "07555", "0788", "-755", "1777", "1000"} {
		require.ErrorIs(t, client.ChmodFile(context.Background(), testServerID, "/", "start.sh", mode), ErrInvalidFileMode)
	}
}

func TestUploadFileUsesScopedTokenAndDoesNotOverwrite(t *testing.T) {
	t.Parallel()

	const (
		userID  = "f1b3d5a7-1234-4567-89ab-cdef01234567"
		content = "uploaded safely"
	)
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch requestCount.Add(1) {
		case 1:
			if got := r.Method; got != http.MethodPost {
				t.Errorf("upload method = %q, want POST", got)
			}
			if got := r.URL.Path; got != "/upload/file" {
				t.Errorf("upload path = %q, want /upload/file", got)
			}
			if got := r.URL.Query().Get("directory"); got != "/worlds/survival" {
				t.Errorf("upload directory = %q, want normalized path", got)
			}
			if got := r.Header.Get("Authorization"); got != "" {
				t.Errorf("upload authorization header = %q, want none", got)
			}
			claims := decodeTestUploadClaims(t, r.URL.Query().Get("token"), "node-token")
			if claims.Scope != "file-upload" || claims.ServerUUID != testServerID || claims.UserUUID != userID || claims.UniqueID == "" {
				t.Errorf("upload claims = %#v, want scoped server/user token", claims)
			}
			if claims.ExpirationTime <= claims.IssuedAt || claims.ExpirationTime-claims.IssuedAt > 60 {
				t.Errorf("upload token lifetime = %d seconds, want at most 60", claims.ExpirationTime-claims.IssuedAt)
			}
			require.NoError(t, r.ParseMultipartForm(1<<20))
			defer r.MultipartForm.RemoveAll()
			files := r.MultipartForm.File["files"]
			require.Len(t, files, 1)
			require.True(t, strings.HasPrefix(files[0].Filename, ".novroforge-upload-"))
			uploaded, err := files[0].Open()
			require.NoError(t, err)
			defer uploaded.Close()
			got, err := io.ReadAll(uploaded)
			require.NoError(t, err)
			require.Equal(t, content, string(got))
			w.WriteHeader(http.StatusOK)
		case 2:
			if got := r.Method; got != http.MethodPut {
				t.Errorf("finalize method = %q, want PUT", got)
			}
			var body struct {
				Root  string `json:"root"`
				Files []struct {
					From string `json:"from"`
					To   string `json:"to"`
				} `json:"files"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
			require.Equal(t, "/worlds/survival", body.Root)
			require.Len(t, body.Files, 1)
			require.True(t, strings.HasPrefix(body.Files[0].From, ".novroforge-upload-"))
			require.Equal(t, "config.json", body.Files[0].To)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request %d: %s %s", requestCount.Load(), r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	require.NoError(t, client.UploadFile(context.Background(), testServerID, userID, "worlds//survival/", "config.json", []byte(content)))
	require.Equal(t, int32(2), requestCount.Load())
}

func TestUploadFileRejectsInvalidInput(t *testing.T) {
	t.Parallel()

	client, err := NewClient("http://localhost:8080", "node-token")
	require.NoError(t, err)
	require.Error(t, client.UploadFile(context.Background(), "invalid", "f1b3d5a7-1234-4567-89ab-cdef01234567", "/", "file.txt", []byte("content")))
	require.Error(t, client.UploadFile(context.Background(), testServerID, "invalid", "/", "file.txt", []byte("content")))
	require.ErrorIs(t, client.UploadFile(context.Background(), testServerID, "f1b3d5a7-1234-4567-89ab-cdef01234567", "../outside", "file.txt", []byte("content")), ErrInvalidDirectory)
	require.ErrorIs(t, client.UploadFile(context.Background(), testServerID, "f1b3d5a7-1234-4567-89ab-cdef01234567", "/", "../file.txt", []byte("content")), ErrInvalidDirectoryName)
	require.ErrorIs(t, client.UploadFile(context.Background(), testServerID, "f1b3d5a7-1234-4567-89ab-cdef01234567", "/", "large.bin", make([]byte, MaxFileUploadSize+1)), ErrFileUploadTooLarge)
}

func TestUploadFileCleansStagingAfterWingsRejectsUpload(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/upload/file":
			w.WriteHeader(http.StatusBadRequest)
		case "/api/servers/" + testServerID + "/files/delete":
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	err = client.UploadFile(context.Background(), testServerID, "f1b3d5a7-1234-4567-89ab-cdef01234567", "/", "file.txt", []byte("content"))
	require.Error(t, err)
	require.Contains(t, err.Error(), "HTTP status 400")
	require.NotContains(t, err.Error(), "token=")
}

func TestUploadFileCleansStagingAfterRequestCancellation(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var cleanupCalled atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/upload/file":
			w.WriteHeader(http.StatusBadRequest)
			flusher, ok := w.(http.Flusher)
			if !ok {
				t.Error("upload response writer does not support flushing")
				return
			}
			flusher.Flush()
			cancel()
		case "/api/servers/" + testServerID + "/files/delete":
			cleanupCalled.Store(true)
			if err := r.Context().Err(); err != nil {
				t.Errorf("cleanup request context error = %v, want active context", err)
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "node-token")
	require.NoError(t, err)
	err = client.UploadFile(ctx, testServerID, "f1b3d5a7-1234-4567-89ab-cdef01234567", "/", "file.txt", []byte("content"))
	require.Error(t, err)
	require.True(t, cleanupCalled.Load(), "staged upload should be removed after cancellation")
}

func decodeTestUploadClaims(t *testing.T, token, secret string) struct {
	IssuedAt       int64  `json:"iat"`
	ExpirationTime int64  `json:"exp"`
	Scope          string `json:"scope"`
	ServerUUID     string `json:"server_uuid"`
	UserUUID       string `json:"user_uuid"`
	UniqueID       string `json:"unique_id"`
} {
	t.Helper()
	parts := strings.Split(token, ".")
	require.Len(t, parts, 3)
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	require.NoError(t, err)
	mac := hmac.New(sha256.New, []byte(secret))
	_, err = mac.Write([]byte(parts[0] + "." + parts[1]))
	require.NoError(t, err)
	require.True(t, hmac.Equal(signature, mac.Sum(nil)))
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	require.NoError(t, err)
	var claims struct {
		IssuedAt       int64  `json:"iat"`
		ExpirationTime int64  `json:"exp"`
		Scope          string `json:"scope"`
		ServerUUID     string `json:"server_uuid"`
		UserUUID       string `json:"user_uuid"`
		UniqueID       string `json:"unique_id"`
	}
	require.NoError(t, json.Unmarshal(payload, &claims))
	return claims
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
