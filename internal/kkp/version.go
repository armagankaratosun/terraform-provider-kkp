package kkp

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "regexp"
    "strings"
)

// GetServerVersion tries a few common KKP version endpoints and returns a version string like "v2.28.2".
// Best-effort: returns non-nil error only for network/HTTP failures; parsing issues return empty version and nil error.
func (c *Client) GetServerVersion(ctx context.Context) (string, error) {
    paths := []string{
        "/api/v2/version",
        "/api/v1/version",
        "/version",
    }

    for _, p := range paths {
        u := *c.BaseURL
        u.Path = strings.TrimRight(u.Path, "/") + p

        req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), http.NoBody)
        if err != nil {
            return "", err
        }
        // headers
        req.Header.Set("User-Agent", defaultUA(""))
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
            continue
        }
        func() { defer func() { _ = resp.Body.Close() }() }()
        if resp.StatusCode >= 400 {
            // try next path
            io.Copy(io.Discard, resp.Body)
            continue
        }
        body, _ := io.ReadAll(resp.Body)
        // Try JSON first
        var m map[string]any
        if json.Unmarshal(body, &m) == nil {
            for _, k := range []string{"gitVersion", "version", "git_version", "git-version"} {
                if v, ok := m[k]; ok {
                    if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
                        return strings.TrimSpace(s), nil
                    }
                }
            }
        }
        // Fallback: plain text like v2.28.2
        s := strings.TrimSpace(string(body))
        if s != "" {
            return s, nil
        }
    }
    return "", fmt.Errorf("version endpoint not found")
}

var mmRe = regexp.MustCompile(`(?i)v?(\d+)\.(\d+)`)

// ExtractMinor returns the major.minor string from a version like v2.28.2 -> 2.28.
func ExtractMinor(v string) string {
    v = strings.TrimSpace(v)
    if v == "" {
        return ""
    }
    m := mmRe.FindStringSubmatch(v)
    if len(m) >= 3 {
        return m[1] + "." + m[2]
    }
    return ""
}

