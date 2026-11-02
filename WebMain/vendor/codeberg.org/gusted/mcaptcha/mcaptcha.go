package mcaptcha

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// VerifyOpts holds all the information that is need to make a verification request.
type VerifyOpts struct {
	Secret      string `json:"secret"`
	Sitekey     string `json:"key"` //nolint:tagliatelle // `Sitekey` is the correct naming, but API expects `key`.
	Token       string `json:"token"`
	InstanceURL string `json:"-"`
}

// GetOpts returns a io.Reader that contains a JSON representation
// of the options.
func (opts *VerifyOpts) GetOpts() (io.Reader, error) {
	if opts.Secret == "" {
		return nil, ErrMissingSecret
	}
	if opts.Sitekey == "" {
		return nil, ErrMissingSitekey
	}
	if opts.Token == "" {
		return nil, ErrMissingToken
	}

	body, err := json.Marshal(opts)
	if err != nil {
		return nil, fmt.Errorf("couldn't marshal options: %w", err)
	}

	return bytes.NewReader(body), nil
}

type verifyResponse struct {
	Valid bool `json:"valid"`
}

// Verify takes in a context and options to make a verification request.
// It will verify if the given token is validated on the given mCaptcha instance.
func Verify(ctx context.Context, opts *VerifyOpts) (bool, error) {
	body, err := opts.GetOpts()
	if err != nil {
		return false, err
	}

	url := strings.TrimSuffix(opts.InstanceURL, "/")

	req, err := http.NewRequestWithContext(ctx, "POST", url+"/api/v1/pow/siteverify", body)
	if err != nil {
		return false, fmt.Errorf("couldn't create a new request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("couldn't execute request: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		content, _ := io.ReadAll(res.Body)
		return false, fmt.Errorf("mCaptcha didn't return 200 OK [content=%q]", string(content))
	}

	var responseStruct verifyResponse
	err = json.NewDecoder(res.Body).Decode(&responseStruct)
	if err != nil {
		return false, fmt.Errorf("couldn't decode response from mCaptcha: %w", err)
	}

	return responseStruct.Valid, nil
}
