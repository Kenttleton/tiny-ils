package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

// enrichVideo fetches movie/show metadata from TMDB (free API key required).
// Set TMDB_API_KEY in the environment. Falls back with an error if not set.
// identifier can be a TMDB numeric ID or a title search string.
func enrichVideo(ctx context.Context, identifier string) (*Result, error) {
	apiKey := os.Getenv("TMDB_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("TMDB_API_KEY not set; video metadata requires a free TMDB API key (https://www.themoviedb.org/settings/api)")
	}

	// Determine if identifier is a numeric TMDB ID
	isTMDBID := isNumeric(identifier)

	var apiURL string
	if isTMDBID {
		apiURL = fmt.Sprintf("https://api.themoviedb.org/3/movie/%s?api_key=%s&language=en-US", identifier, apiKey)
	} else {
		apiURL = fmt.Sprintf("https://api.themoviedb.org/3/search/movie?api_key=%s&query=%s&page=1", apiKey, url.QueryEscape(identifier))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tmdb request: %w", err)
	}
	defer resp.Body.Close()

	if isTMDBID {
		return parseTMDBMovie(resp)
	}
	return parseTMDBSearch(resp)
}

func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

func parseTMDBMovie(resp *http.Response) (*Result, error) {
	var raw struct {
		Title    string `json:"title"`
		Overview string `json:"overview"`
		Genres   []struct {
			Name string `json:"name"`
		} `json:"genres"`
		PosterPath string `json:"poster_path"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode tmdb: %w", err)
	}
	r := &Result{Title: raw.Title, Description: raw.Overview, Source: "themoviedb.org"}
	for _, g := range raw.Genres {
		r.Tags = append(r.Tags, g.Name)
	}
	if raw.PosterPath != "" {
		r.CoverURL = "https://image.tmdb.org/t/p/w500" + raw.PosterPath
	}
	return r, nil
}

func parseTMDBSearch(resp *http.Response) (*Result, error) {
	var raw struct {
		Results []struct {
			Title    string `json:"title"`
			Overview string `json:"overview"`
			PosterPath string `json:"poster_path"`
		} `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode tmdb search: %w", err)
	}
	if len(raw.Results) == 0 {
		return nil, fmt.Errorf("no results from TMDB")
	}
	m := raw.Results[0]
	r := &Result{Title: m.Title, Description: m.Overview, Source: "themoviedb.org"}
	if m.PosterPath != "" {
		r.CoverURL = "https://image.tmdb.org/t/p/w500" + m.PosterPath
	}
	return r, nil
}
