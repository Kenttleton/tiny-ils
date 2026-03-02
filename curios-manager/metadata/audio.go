package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// enrichAudio fetches music metadata from MusicBrainz (no API key required, FOSS).
// identifier can be a MusicBrainz release MBID or a search string.
func enrichAudio(ctx context.Context, identifier string) (*Result, error) {
	isMBID := len(identifier) == 36 && identifier[8] == '-'

	var apiURL string
	if isMBID {
		apiURL = "https://musicbrainz.org/ws/2/release/" + identifier + "?inc=artists+genres&fmt=json"
	} else {
		apiURL = "https://musicbrainz.org/ws/2/release?query=" + url.QueryEscape(identifier) + "&limit=1&fmt=json"
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, err
	}
	// MusicBrainz requires a descriptive User-Agent
	req.Header.Set("User-Agent", "tiny-ils/1.0 (https://github.com/tiny-ils)")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("musicbrainz request: %w", err)
	}
	defer resp.Body.Close()

	if isMBID {
		return parseMBRelease(resp)
	}
	return parseMBSearch(resp)
}

func parseMBRelease(resp *http.Response) (*Result, error) {
	var raw struct {
		Title   string `json:"title"`
		Artists []struct {
			Artist struct {
				Name string `json:"name"`
			} `json:"artist"`
		} `json:"artist-credit"`
		Genres []struct {
			Name string `json:"name"`
		} `json:"genres"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode musicbrainz: %w", err)
	}
	r := &Result{Title: raw.Title, Source: "musicbrainz.org"}
	for _, a := range raw.Artists {
		r.Authors = append(r.Authors, a.Artist.Name)
	}
	for _, g := range raw.Genres {
		r.Tags = append(r.Tags, g.Name)
	}
	return r, nil
}

func parseMBSearch(resp *http.Response) (*Result, error) {
	var raw struct {
		Releases []struct {
			Title   string `json:"title"`
			Artists []struct {
				Artist struct {
					Name string `json:"name"`
				} `json:"artist"`
			} `json:"artist-credit"`
		} `json:"releases"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode musicbrainz search: %w", err)
	}
	if len(raw.Releases) == 0 {
		return nil, fmt.Errorf("no results from MusicBrainz")
	}
	rel := raw.Releases[0]
	r := &Result{Title: rel.Title, Source: "musicbrainz.org"}
	for _, a := range rel.Artists {
		r.Authors = append(r.Authors, a.Artist.Name)
	}
	return r, nil
}
