// Package kkp provides a KKP client wrapper and authentication utilities.
package kkp

import (
	"context"
	"io"
	"net/http"

	runtime "github.com/go-openapi/runtime"
	httptransport "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
	gok "github.com/kubermatic/go-kubermatic/client"
)

// AuthenticateRequest implements the runtime.ClientAuthInfoWriter interface.
// It sets the Authorization header with the Bearer token, User-Agent, and any extra headers
func (a *headerAuthWriter) AuthenticateRequest(rq runtime.ClientRequest, _ strfmt.Registry) error {
	if a.userAgent != "" {
		_ = rq.SetHeaderParam("User-Agent", a.userAgent)
	}
	for k, v := range a.extraHeader {
		_ = rq.SetHeaderParam(k, v)
	}
	if a.token != "" {
		_ = rq.SetHeaderParam("Authorization", "Bearer "+a.token)
	}
	return nil
}

// NewHTTPClient creates a configured KKP REST client.
func NewHTTPClient(cfg Config) (*Client, error) {
	baseURL, err := normalizeBase(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	tlsCfg, err := buildTLSConfig(cfg.InsecureSkipVerify, cfg.CAFile)
	if err != nil {
		return nil, err
	}
	httpClient := newHTTPClient(cfg.Timeout, tlsCfg)

	rt := httptransport.NewWithClient(baseURL.Host, baseURL.Path, []string{baseURL.Scheme}, httpClient)
	rt.DefaultAuthentication = &headerAuthWriter{
		token:       cfg.Token,
		userAgent:   defaultUA(cfg.UserAgent),
		extraHeader: cfg.ExtraHeaders,
	}

	api := gok.New(rt, strfmt.Default)

	return &Client{
		API:        api, // keep as `any` unless you pin the exact type
		Transport:  rt,
		HTTPClient: httpClient,
		BaseURL:    baseURL,
	}, nil
}

// Ping checks the base API path for reachability (2xx/3xx).
func (c *Client) Ping(ctx context.Context) error {
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL.String(), http.NoBody)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", defaultUA(""))

	// Reuse headers from our auth writer if present.
	if aw, ok := c.Transport.DefaultAuthentication.(*headerAuthWriter); ok {
		if aw.token != "" {
			req.Header.Set("Authorization", "Bearer "+aw.token)
		}
		for k, v := range aw.extraHeader {
			req.Header.Set(k, v)
		}
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode >= 400 {
		return &httpError{status: resp.Status}
	}
	return nil
}

type httpError struct{ status string }

func (e *httpError) Error() string { return "kkp ping: " + e.status }
