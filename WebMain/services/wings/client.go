// Copyright 2026 The NovroForge Authors. All rights reserved.
// SPDX-License-Identifier: GPL-3.0-or-later

package wings

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	maxResponseSize    = 4 << 20
	maxTokenLength     = 4096
	maxCommands        = 32
	maxCommandLength   = 4096
	maxDirectoryPath   = 4096
	// MaxFileContentSize bounds text that can be previewed or edited.
	MaxFileContentSize = 1 << 20
	// MaxFileUploadSize bounds files uploaded through the WebMain panel.
	MaxFileUploadSize = 10 << 20
	maxFilePreviewSize  = MaxFileContentSize
	maxFileNameLength  = 255
)

var (
	ErrInvalidDirectory    = errors.New("invalid Wings directory path")
	ErrInvalidFilePath     = errors.New("invalid Wings file path")
	ErrInvalidDirectoryName = errors.New("invalid Wings directory name")
	ErrFileTooLarge        = errors.New("Wings file exceeds the preview size limit")
	ErrFileUploadTooLarge   = errors.New("Wings file exceeds the upload size limit")
	ErrUnsupportedFileType  = errors.New("Wings file is not a supported text type")
)

// Client sends authenticated requests to a Wings daemon.
type Client struct {
	baseURL    *url.URL
	token      string
	httpClient *http.Client
}

// NewClient creates a client for a Wings node. Plain HTTP is only accepted for
// loopback addresses so node credentials cannot be sent over an untrusted link.
func NewClient(nodeURL, token string) (*Client, error) {
	if len(token) == 0 || len(token) > maxTokenLength {
		return nil, errors.New("Wings token length is invalid")
	}
	for _, char := range token {
		if char < '!' || char > '~' {
			return nil, errors.New("Wings token must contain only visible ASCII characters")
		}
	}

	baseURL, err := url.Parse(strings.TrimSpace(nodeURL))
	if err != nil {
		return nil, fmt.Errorf("invalid Wings URL: %w", err)
	}
	if baseURL.Hostname() == "" || baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" || (baseURL.Path != "" && baseURL.Path != "/") {
		return nil, errors.New("Wings URL must contain only a host and optional port")
	}
	if baseURL.Scheme != "https" && baseURL.Scheme != "http" {
		return nil, errors.New("Wings URL must use HTTPS")
	}
	if baseURL.Scheme == "http" {
		hostname := strings.TrimSuffix(strings.ToLower(baseURL.Hostname()), ".")
		ip := net.ParseIP(hostname)
		if hostname != "localhost" && (ip == nil || !ip.IsLoopback()) {
			return nil, errors.New("Wings URL must use HTTPS unless it targets loopback")
		}
	}
	baseURL.Path = ""
	baseURL.RawPath = ""

	return &Client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}, nil
}

// Utilization contains the current resource usage reported by Wings.
type Utilization struct {
	MemoryBytes uint64  `json:"memory_bytes"`
	MemoryLimit uint64  `json:"memory_limit_bytes"`
	CPUAbsolute float64 `json:"cpu_absolute"`
	DiskBytes   int64   `json:"disk_bytes"`
	Uptime      int64   `json:"uptime"`
	Network     struct {
		RxBytes int64 `json:"rx_bytes"`
		TxBytes int64 `json:"tx_bytes"`
	} `json:"network"`
}

// ServerDetails is the state and resource information returned by Wings.
type ServerDetails struct {
	State       string      `json:"state"`
	IsSuspended bool        `json:"is_suspended"`
	Utilization Utilization `json:"utilization"`
}

// FileEntry describes one item in a Wings server directory.
type FileEntry struct {
	Name      string `json:"name"`
	Created   string `json:"created"`
	Modified  string `json:"modified"`
	Mode      string `json:"mode"`
	ModeBits  string `json:"mode_bits"`
	Size      int64  `json:"size"`
	Directory bool   `json:"directory"`
	File      bool   `json:"file"`
	Symlink   bool   `json:"symlink"`
	Mime      string `json:"mime"`
}

// SystemInformation is returned by Wings when testing a node connection.
type SystemInformation struct {
	Architecture  string `json:"architecture"`
	CPUCount      int    `json:"cpu_count"`
	KernelVersion string `json:"kernel_version"`
	OS            string `json:"os"`
	Version       string `json:"version"`
}

// GetSystem retrieves daemon information and verifies node connectivity.
func (c *Client) GetSystem(ctx context.Context) (SystemInformation, error) {
	var info SystemInformation
	err := c.doJSON(ctx, http.MethodGet, "/api/system", nil, &info, http.StatusOK)
	return info, err
}

// GetServer retrieves the state and resource usage for a server.
func (c *Client) GetServer(ctx context.Context, serverID string) (ServerDetails, error) {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return ServerDetails{}, errors.New("invalid Wings server UUID")
	}

	var details ServerDetails
	err = c.doJSON(ctx, http.MethodGet, "/api/servers/"+id.String(), nil, &details, http.StatusOK)
	return details, err
}

// Logs retrieves up to 100 recent console lines for a server.
func (c *Client) Logs(ctx context.Context, serverID string, lines int) ([]string, error) {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return nil, errors.New("invalid Wings server UUID")
	}
	if lines < 1 || lines > 100 {
		return nil, errors.New("log line count must be between 1 and 100")
	}

	var response struct {
		Data []string `json:"data"`
	}
	err = c.doJSON(ctx, http.MethodGet, "/api/servers/"+id.String()+"/logs?size="+strconv.Itoa(lines), nil, &response, http.StatusOK)
	return response.Data, err
}

// ListDirectory retrieves metadata for entries in a server directory.
func (c *Client) ListDirectory(ctx context.Context, serverID, directory string) ([]FileEntry, error) {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return nil, errors.New("invalid Wings server UUID")
	}
	directory, err = NormalizeDirectoryPath(directory)
	if err != nil {
		return nil, err
	}

	requestPath := "/api/servers/" + id.String() + "/files/list-directory?" + url.Values{
		"directory": []string{directory},
	}.Encode()
	var entries []FileEntry
	err = c.doJSON(ctx, http.MethodGet, requestPath, nil, &entries, http.StatusOK)
	return entries, err
}

// NormalizeDirectoryPath validates and normalizes a path within a server root.
func NormalizeDirectoryPath(directory string) (string, error) {
	if len(directory) > maxDirectoryPath || strings.ContainsAny(directory, "\\\x00") {
		return "", ErrInvalidDirectory
	}
	for _, segment := range strings.Split(directory, "/") {
		if segment == ".." {
			return "", ErrInvalidDirectory
		}
	}
	if directory == "" {
		return "/", nil
	}
	return path.Clean("/" + strings.TrimLeft(directory, "/")), nil
}

// FileContents retrieves a bounded UTF-8 text preview from a server file.
func (c *Client) FileContents(ctx context.Context, serverID, filePath string) (string, error) {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return "", errors.New("invalid Wings server UUID")
	}
	filePath, err = normalizeFilePath(filePath)
	if err != nil || filePath == "/" {
		return "", ErrInvalidFilePath
	}

	requestPath := "/api/servers/" + id.String() + "/files/contents?" + url.Values{
		"file": []string{filePath},
	}.Encode()
	response, err := c.doRequest(ctx, http.MethodGet, requestPath, nil, http.StatusOK)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	contentType, _, err := mime.ParseMediaType(response.Header.Get("X-Mime-Type"))
	if err != nil || !isSupportedTextMime(contentType) {
		return "", ErrUnsupportedFileType
	}
	if response.ContentLength > maxFilePreviewSize {
		return "", ErrFileTooLarge
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, maxFilePreviewSize+1))
	if err != nil {
		return "", fmt.Errorf("read Wings file preview: %w", err)
	}
	if len(content) > maxFilePreviewSize {
		return "", ErrFileTooLarge
	}
	if !utf8.Valid(content) {
		return "", ErrUnsupportedFileType
	}
	return string(content), nil
}

// WriteFile replaces a bounded UTF-8 text file through the Wings file API.
func (c *Client) WriteFile(ctx context.Context, serverID, filePath, content string) error {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return errors.New("invalid Wings server UUID")
	}
	filePath, err = normalizeFilePath(filePath)
	if err != nil || filePath == "/" {
		return ErrInvalidFilePath
	}
	if len(content) > maxFilePreviewSize {
		return ErrFileTooLarge
	}
	if !utf8.ValidString(content) {
		return ErrUnsupportedFileType
	}

	requestPath := "/api/servers/" + id.String() + "/files/write?" + url.Values{
		"file": []string{filePath},
	}.Encode()
	response, err := c.doRequest(ctx, http.MethodPost, requestPath, []byte(content), http.StatusNoContent)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	return nil
}

// FileDownload opens a file response for streaming. The caller must close the
// response body; request cancellation controls the lifetime of the download.
func (c *Client) FileDownload(ctx context.Context, serverID, filePath string) (*http.Response, error) {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return nil, errors.New("invalid Wings server UUID")
	}
	filePath, err = normalizeFilePath(filePath)
	if err != nil || filePath == "/" {
		return nil, ErrInvalidFilePath
	}

	requestPath := "/api/servers/" + id.String() + "/files/contents?" + url.Values{
		"file": []string{filePath},
	}.Encode()
	downloadClient := *c.httpClient
	downloadClient.Timeout = 0
	return c.doRequestWithClient(ctx, &downloadClient, http.MethodGet, requestPath, nil, http.StatusOK)
}

// CreateDirectory creates a new directory without replacing existing files.
func (c *Client) CreateDirectory(ctx context.Context, serverID, name, directory string) error {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return errors.New("invalid Wings server UUID")
	}
	if !isValidDirectoryName(name) {
		return ErrInvalidDirectoryName
	}
	directory, err = NormalizeDirectoryPath(directory)
	if err != nil {
		return err
	}

	body := struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}{Name: name, Path: directory}
	return c.doJSON(ctx, http.MethodPost, "/api/servers/"+id.String()+"/files/create-directory", body, nil, http.StatusNoContent)
}

// RenameFile renames one entry within its current directory. Wings rejects an
// existing destination instead of replacing it.
func (c *Client) RenameFile(ctx context.Context, serverID, directory, from, to string) error {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return errors.New("invalid Wings server UUID")
	}
	directory, err = NormalizeDirectoryPath(directory)
	if err != nil {
		return err
	}
	if !isValidDirectoryName(from) || !isValidDirectoryName(to) || from == to {
		return ErrInvalidDirectoryName
	}

	body := struct {
		Root  string `json:"root"`
		Files []struct {
			From string `json:"from"`
			To   string `json:"to"`
		} `json:"files"`
	}{Root: directory, Files: []struct {
		From string `json:"from"`
		To   string `json:"to"`
	}{{From: from, To: to}}}
	return c.doJSON(ctx, http.MethodPut, "/api/servers/"+id.String()+"/files/rename", body, nil, http.StatusNoContent)
}

// DeleteFile deletes one named entry under its parent directory.
func (c *Client) DeleteFile(ctx context.Context, serverID, directory, name string) error {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return errors.New("invalid Wings server UUID")
	}
	directory, err = NormalizeDirectoryPath(directory)
	if err != nil {
		return err
	}
	if !isValidDirectoryName(name) {
		return ErrInvalidDirectoryName
	}

	body := struct {
		Root  string   `json:"root"`
		Files []string `json:"files"`
	}{Root: directory, Files: []string{name}}
	return c.doJSON(ctx, http.MethodPost, "/api/servers/"+id.String()+"/files/delete", body, nil, http.StatusNoContent)
}

// CopyFile creates a copy of one regular file in the same directory.
func (c *Client) CopyFile(ctx context.Context, serverID, filePath string) error {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return errors.New("invalid Wings server UUID")
	}
	filePath, err = normalizeFilePath(filePath)
	if err != nil || filePath == "/" {
		return ErrInvalidFilePath
	}

	body := struct {
		Location string `json:"location"`
	}{Location: filePath}
	return c.doJSON(ctx, http.MethodPost, "/api/servers/"+id.String()+"/files/copy", body, nil, http.StatusNoContent)
}

// CompressFile creates a tar.gz archive containing one file in its directory.
func (c *Client) CompressFile(ctx context.Context, serverID, directory, name string) (FileEntry, error) {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return FileEntry{}, errors.New("invalid Wings server UUID")
	}
	directory, err = NormalizeDirectoryPath(directory)
	if err != nil {
		return FileEntry{}, err
	}
	if !isValidDirectoryName(name) {
		return FileEntry{}, ErrInvalidDirectoryName
	}

	body := struct {
		Root  string   `json:"root"`
		Files []string `json:"files"`
	}{Root: directory, Files: []string{name}}
	var archive FileEntry
	compressClient := *c.httpClient
	compressClient.Timeout = 0
	err = c.doJSONWithClient(ctx, &compressClient, http.MethodPost, "/api/servers/"+id.String()+"/files/compress", body, &archive, http.StatusOK)
	return archive, err
}

// UploadFile uploads one bounded file using a short-lived, single-use Wings token.
// The file is staged under a random name and renamed so an existing destination
// cannot be overwritten.
func (c *Client) UploadFile(ctx context.Context, serverID, userID, directory, name string, content []byte) error {
	serverUUID, err := uuid.Parse(serverID)
	if err != nil {
		return errors.New("invalid Wings server UUID")
	}
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return errors.New("invalid Wings user UUID")
	}
	directory, err = NormalizeDirectoryPath(directory)
	if err != nil {
		return err
	}
	if !isValidDirectoryName(name) {
		return ErrInvalidDirectoryName
	}
	if len(content) > MaxFileUploadSize {
		return ErrFileUploadTooLarge
	}

	stagingName := ".novroforge-upload-" + uuid.NewString() + ".tmp"
	token, err := c.createFileUploadToken(serverUUID.String(), userUUID.String())
	if err != nil {
		return fmt.Errorf("create Wings upload token: %w", err)
	}
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("files", stagingName)
	if err != nil {
		return fmt.Errorf("create Wings upload part: %w", err)
	}
	if _, err := part.Write(content); err != nil {
		return fmt.Errorf("write Wings upload part: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish Wings upload body: %w", err)
	}

	requestURL := *c.baseURL
	requestURL.Path = "/upload/file"
	requestURL.RawPath = ""
	requestURL.RawQuery = url.Values{
		"directory": []string{directory},
		"token":     []string{token},
	}.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), &body)
	if err != nil {
		return errors.New("create Wings upload request failed")
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	uploadClient := *c.httpClient
	uploadClient.Timeout = 0
	response, err := uploadClient.Do(request)
	if err != nil {
		requestErr := errors.New("upload file to Wings request failed")
		if ctxErr := ctx.Err(); ctxErr != nil {
			requestErr = fmt.Errorf("upload file to Wings request failed: %w", ctxErr)
		}
		return cleanupWingsUpload(ctx, c, serverUUID.String(), directory, stagingName, requestErr)
	}
	closeErr := response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return cleanupWingsUpload(ctx, c, serverUUID.String(), directory, stagingName, fmt.Errorf("Wings file upload returned unexpected HTTP status %d", response.StatusCode))
	}
	if closeErr != nil {
		return cleanupWingsUpload(ctx, c, serverUUID.String(), directory, stagingName, errors.New("close Wings upload response failed"))
	}

	if err := c.RenameFile(ctx, serverUUID.String(), directory, stagingName, name); err != nil {
		return cleanupWingsUpload(ctx, c, serverUUID.String(), directory, stagingName, fmt.Errorf("finalize Wings file upload: %w", err))
	}
	return nil
}

func cleanupWingsUpload(ctx context.Context, client *Client, serverID, directory, stagingName string, cause error) error {
	if cleanupErr := client.DeleteFile(ctx, serverID, directory, stagingName); cleanupErr != nil {
		return fmt.Errorf("%w (temporary file cleanup failed: %v)", cause, cleanupErr)
	}
	return cause
}

func (c *Client) createFileUploadToken(serverID, userID string) (string, error) {
	now := time.Now()
	uniqueID := uuid.NewString()
	payload := struct {
		IssuedAt       int64  `json:"iat"`
		ExpirationTime int64  `json:"exp"`
		JWTID          string `json:"jti"`
		Scope          string `json:"scope"`
		ServerUUID     string `json:"server_uuid"`
		UserUUID       string `json:"user_uuid"`
		UniqueID       string `json:"unique_id"`
	}{
		IssuedAt:       now.Add(time.Second).Unix(), // Wings rejects tokens issued before its boot time.
		ExpirationTime: now.Add(time.Minute).Unix(),
		JWTID:          uniqueID,
		Scope:          "file-upload",
		ServerUUID:     serverID,
		UserUUID:       userID,
		UniqueID:       uniqueID,
	}
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("encode upload claims: %w", err)
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	encodedPayload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	unsignedToken := header + "." + encodedPayload
	signature := hmac.New(sha256.New, []byte(c.token))
	if _, err := signature.Write([]byte(unsignedToken)); err != nil {
		return "", fmt.Errorf("sign upload claims: %w", err)
	}
	return unsignedToken + "." + base64.RawURLEncoding.EncodeToString(signature.Sum(nil)), nil
}

func isValidDirectoryName(name string) bool {
	if name == "" || name == "." || name == ".." || len(name) > maxFileNameLength || !utf8.ValidString(name) || strings.TrimSpace(name) != name || strings.ContainsAny(name, "/\\\x00") {
		return false
	}
	for _, char := range name {
		if char < 0x20 || char == 0x7f {
			return false
		}
	}
	return true
}

// ValidateFileName validates one file or directory entry name.
func ValidateFileName(name string) error {
	if !isValidDirectoryName(name) {
		return ErrInvalidDirectoryName
	}
	return nil
}

func normalizeFilePath(filePath string) (string, error) {
	if strings.HasSuffix(filePath, "/") {
		return "", ErrInvalidFilePath
	}
	return NormalizeDirectoryPath(filePath)
}

func isSupportedTextMime(contentType string) bool {
	if strings.HasPrefix(contentType, "text/") {
		return true
	}
	switch contentType {
	case "application/json", "application/javascript", "application/xml", "application/x-yaml", "application/toml":
		return true
	default:
		return false
	}
}

// Power sends a power action to a server.
func (c *Client) Power(ctx context.Context, serverID, action string) error {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return errors.New("invalid Wings server UUID")
	}
	switch action {
	case "start", "stop", "restart", "kill":
	default:
		return errors.New("unsupported Wings power action")
	}

	body := struct {
		Action string `json:"action"`
	}{Action: action}
	return c.doJSON(ctx, http.MethodPost, "/api/servers/"+id.String()+"/power", body, nil, http.StatusAccepted)
}

// Commands sends commands to a server console.
func (c *Client) Commands(ctx context.Context, serverID string, commands []string) error {
	id, err := uuid.Parse(serverID)
	if err != nil {
		return errors.New("invalid Wings server UUID")
	}
	if len(commands) == 0 || len(commands) > maxCommands {
		return fmt.Errorf("server command count must be between 1 and %d", maxCommands)
	}
	for _, command := range commands {
		if strings.TrimSpace(command) == "" || len(command) > maxCommandLength {
			return fmt.Errorf("server commands must be non-empty and no longer than %d bytes", maxCommandLength)
		}
	}

	body := struct {
		Commands []string `json:"commands"`
	}{Commands: commands}
	return c.doJSON(ctx, http.MethodPost, "/api/servers/"+id.String()+"/commands", body, nil, http.StatusNoContent)
}

func (c *Client) doJSON(ctx context.Context, method, requestPath string, body, result any, expectedStatus int) error {
	return c.doJSONWithClient(ctx, c.httpClient, method, requestPath, body, result, expectedStatus)
}

func (c *Client) doJSONWithClient(ctx context.Context, client *http.Client, method, requestPath string, body, result any, expectedStatus int) error {
	response, err := c.doRequestWithClient(ctx, client, method, requestPath, body, expectedStatus)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if result == nil {
		return nil
	}

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, maxResponseSize+1))
	if err != nil {
		return fmt.Errorf("read Wings response: %w", err)
	}
	if len(responseBody) > maxResponseSize {
		return errors.New("Wings response exceeded the maximum allowed size")
	}
	if err := json.Unmarshal(responseBody, result); err != nil {
		return fmt.Errorf("decode Wings response: %w", err)
	}
	return nil
}

func (c *Client) doRequest(ctx context.Context, method, requestPath string, body any, expectedStatus int) (*http.Response, error) {
	return c.doRequestWithClient(ctx, c.httpClient, method, requestPath, body, expectedStatus)
}

func (c *Client) doRequestWithClient(ctx context.Context, client *http.Client, method, requestPath string, body any, expectedStatus int) (*http.Response, error) {
	var requestBody io.Reader
	contentType := ""
	if rawBody, ok := body.([]byte); ok {
		requestBody = bytes.NewReader(rawBody)
		contentType = "application/octet-stream"
	} else if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode Wings request: %w", err)
		}
		requestBody = bytes.NewReader(encoded)
		contentType = "application/json"
	}

	requestURL := *c.baseURL
	parsedPath, err := url.Parse(requestPath)
	if err != nil || parsedPath.IsAbs() || parsedPath.Host != "" {
		return nil, errors.New("invalid Wings API path")
	}
	requestURL = *requestURL.ResolveReference(parsedPath)
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), requestBody)
	if err != nil {
		return nil, fmt.Errorf("create Wings request: %w", err)
	}
	request.Header.Set("Authorization", "Bearer "+c.token)
	request.Header.Set("Accept", "application/json")
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request to Wings failed: %w", err)
	}
	if response.StatusCode != expectedStatus {
		response.Body.Close()
		return nil, fmt.Errorf("Wings returned unexpected HTTP status %d", response.StatusCode)
	}
	return response, nil
}
