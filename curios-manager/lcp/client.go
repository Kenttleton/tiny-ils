// Package lcp provides an HTTP client for the Readium LCP server (lcpserver)
// and License Status Document server (lsdserver).
package lcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client talks to lcpserver and lsdserver.
type Client struct {
	lcpURL string
	lsdURL string
	http   *http.Client
}

// NewClient creates a Client for the given server base URLs.
// lcpURL example: "http://lcp-server:10000"
// lsdURL example: "http://lsd-server:10001"
func NewClient(lcpURL, lsdURL string) *Client {
	return &Client{
		lcpURL: lcpURL,
		lsdURL: lsdURL,
		http:   &http.Client{Timeout: 30 * time.Second},
	}
}

// ─── License issuance ────────────────────────────────────────────────────────

// LicenseRequest is the payload sent to lcpserver POST /api/v1/licenses.
type LicenseRequest struct {
	Provider  string        `json:"provider"`
	User      LicenseUser   `json:"user"`
	Rights    LicenseRights `json:"rights"`
	ContentID string        `json:"content_id"`
}

// LicenseUser identifies the patron in the LCP license.
type LicenseUser struct {
	ID string `json:"id"`
}

// LicenseRights describes the access window for the licensed content.
type LicenseRights struct {
	Start *time.Time `json:"start,omitempty"`
	End   *time.Time `json:"end,omitempty"`
}

// LicenseDocument is the LCP License JSON returned by lcpserver.
type LicenseDocument struct {
	ID       string         `json:"id"`
	Provider string         `json:"provider"`
	Issued   time.Time      `json:"issued"`
	Rights   LicenseRights  `json:"rights,omitempty"`
	Links    []LicenseLink  `json:"links,omitempty"`
}

// LicenseLink is a link embedded in the LCP License document.
type LicenseLink struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
	Type string `json:"type,omitempty"`
}

// IssueLicense calls lcpserver POST /api/v1/licenses and returns the signed
// LCP License document.
func (c *Client) IssueLicense(ctx context.Context, req LicenseRequest) (*LicenseDocument, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("lcp: marshal license request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx, http.MethodPost, c.lcpURL+"/api/v1/licenses", bytes.NewReader(body),
	)
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("lcp: issue license: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("lcp: issue license: status %d: %s", resp.StatusCode, string(b))
	}

	var license LicenseDocument
	if err := json.NewDecoder(resp.Body).Decode(&license); err != nil {
		return nil, fmt.Errorf("lcp: decode license: %w", err)
	}
	return &license, nil
}

// ─── Content registration ─────────────────────────────────────────────────────

// RegisterContent uploads an encrypted content file to lcpserver via
// PUT /content/{contentID}.  The caller is responsible for encrypting
// the file before calling this method (using lcpencrypt or equivalent).
func (c *Client) RegisterContent(ctx context.Context, contentID string, body io.Reader, mediaType string) error {
	url := fmt.Sprintf("%s/content/%s", c.lcpURL, contentID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, body)
	if err != nil {
		return err
	}
	if mediaType != "" {
		req.Header.Set("Content-Type", mediaType)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("lcp: register content: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("lcp: register content: status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}

// ─── License revocation ───────────────────────────────────────────────────────

// ReturnLicense calls lsdserver POST /licenses/{id}/return to mark the
// license as returned.  Thorium Reader detects this on its next sync.
func (c *Client) ReturnLicense(ctx context.Context, licenseID string) error {
	url := fmt.Sprintf("%s/licenses/%s/return", c.lsdURL, licenseID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("lsd: return license: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("lsd: return license: status %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
