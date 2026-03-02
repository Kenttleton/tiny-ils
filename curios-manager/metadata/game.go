package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	igdbTokenMu    sync.Mutex
	igdbToken      string
	igdbTokenExpiry time.Time
)

// enrichGame fetches game metadata from IGDB (free via Twitch OAuth).
// Set IGDB_CLIENT_ID and IGDB_CLIENT_SECRET in the environment.
func enrichGame(ctx context.Context, identifier string) (*Result, error) {
	clientID := os.Getenv("IGDB_CLIENT_ID")
	clientSecret := os.Getenv("IGDB_CLIENT_SECRET")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("IGDB_CLIENT_ID and IGDB_CLIENT_SECRET not set; game metadata requires free IGDB credentials (https://api-docs.igdb.com/#getting-started)")
	}

	token, err := getIGDBToken(ctx, clientID, clientSecret)
	if err != nil {
		return nil, err
	}

	isID := isNumeric(identifier)
	var body string
	if isID {
		body = fmt.Sprintf("fields name,summary,genres.name,cover.url; where id = %s;", identifier)
	} else {
		body = fmt.Sprintf(`fields name,summary,genres.name,cover.url; search "%s"; limit 1;`, identifier)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.igdb.com/v4/games", strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-ID", clientID)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "text/plain")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("igdb request: %w", err)
	}
	defer resp.Body.Close()

	var games []struct {
		Name    string `json:"name"`
		Summary string `json:"summary"`
		Genres  []struct {
			Name string `json:"name"`
		} `json:"genres"`
		Cover struct {
			URL string `json:"url"`
		} `json:"cover"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&games); err != nil {
		return nil, fmt.Errorf("decode igdb: %w", err)
	}
	if len(games) == 0 {
		return nil, fmt.Errorf("no results from IGDB")
	}
	g := games[0]
	r := &Result{Title: g.Name, Description: g.Summary, Source: "igdb.com"}
	for _, genre := range g.Genres {
		r.Tags = append(r.Tags, genre.Name)
	}
	if g.Cover.URL != "" {
		r.CoverURL = "https:" + strings.Replace(g.Cover.URL, "t_thumb", "t_cover_big", 1)
	}
	return r, nil
}

func getIGDBToken(ctx context.Context, clientID, clientSecret string) (string, error) {
	igdbTokenMu.Lock()
	defer igdbTokenMu.Unlock()

	if igdbToken != "" && time.Now().Before(igdbTokenExpiry) {
		return igdbToken, nil
	}

	resp, err := http.PostForm("https://id.twitch.tv/oauth2/token", url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"grant_type":    {"client_credentials"},
	})
	if err != nil {
		return "", fmt.Errorf("igdb token request: %w", err)
	}
	defer resp.Body.Close()

	var tok struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tok); err != nil {
		return "", fmt.Errorf("decode igdb token: %w", err)
	}
	igdbToken = tok.AccessToken
	igdbTokenExpiry = time.Now().Add(time.Duration(tok.ExpiresIn-60) * time.Second)
	return igdbToken, nil
}
