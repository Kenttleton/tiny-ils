package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// enrichBook fetches book metadata from Open Library (no API key required).
// identifier can be an ISBN (10 or 13) or a title search string.
func enrichBook(ctx context.Context, identifier string) (*Result, error) {
	var apiURL string
	isISBN := isISBN(identifier)
	if isISBN {
		apiURL = "https://openlibrary.org/api/books?bibkeys=ISBN:" + identifier + "&format=json&jscmd=data"
	} else {
		apiURL = "https://openlibrary.org/search.json?title=" + url.QueryEscape(identifier) + "&limit=1"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "tiny-ils/1.0 (https://github.com/tiny-ils)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("open library request: %w", err)
	}
	defer resp.Body.Close()

	if isISBN {
		return parseOpenLibraryByISBN(resp, identifier)
	}
	return parseOpenLibrarySearch(resp)
}

func isISBN(s string) bool {
	clean := strings.ReplaceAll(s, "-", "")
	return len(clean) == 10 || len(clean) == 13
}

func parseOpenLibraryByISBN(resp *http.Response, isbn string) (*Result, error) {
	var raw map[string]struct {
		Title   string `json:"title"`
		Authors []struct {
			Name string `json:"name"`
		} `json:"authors"`
		Cover struct {
			Large string `json:"large"`
		} `json:"cover"`
		Subjects []struct {
			Name string `json:"name"`
		} `json:"subjects"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode open library: %w", err)
	}

	book, ok := raw["ISBN:"+isbn]
	if !ok {
		return nil, fmt.Errorf("ISBN %s not found on Open Library", isbn)
	}

	r := &Result{Title: book.Title, Source: "openlibrary.org"}
	for _, a := range book.Authors {
		r.Authors = append(r.Authors, a.Name)
	}
	r.CoverURL = book.Cover.Large
	for _, s := range book.Subjects {
		r.Tags = append(r.Tags, s.Name)
	}
	return r, nil
}

func parseOpenLibrarySearch(resp *http.Response) (*Result, error) {
	var raw struct {
		Docs []struct {
			Title         string   `json:"title"`
			AuthorName    []string `json:"author_name"`
			Subject       []string `json:"subject"`
			CoverI        int      `json:"cover_i"`
		} `json:"docs"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode open library search: %w", err)
	}
	if len(raw.Docs) == 0 {
		return nil, fmt.Errorf("no results from Open Library")
	}
	d := raw.Docs[0]
	r := &Result{
		Title:   d.Title,
		Authors: d.AuthorName,
		Tags:    d.Subject,
		Source:  "openlibrary.org",
	}
	if d.CoverI != 0 {
		r.CoverURL = fmt.Sprintf("https://covers.openlibrary.org/b/id/%d-L.jpg", d.CoverI)
	}
	return r, nil
}
