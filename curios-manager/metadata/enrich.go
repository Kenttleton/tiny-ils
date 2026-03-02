// Package metadata fetches curio metadata from free/open external APIs.
// THING types are always manual; all other types attempt enrichment and fall
// back gracefully if the API is unavailable or no key is configured.
package metadata

import (
	"context"
	"fmt"

	"tiny-ils/shared/models"
)

type Result struct {
	Title       string
	Description string
	Authors     []string
	CoverURL    string
	Tags        []string
	Source      string
}

// Enrich looks up metadata for the given media type and identifier.
// identifier is: ISBN for books, MusicBrainz release ID for audio,
// TMDB ID for video, IGDB ID for games.
func Enrich(ctx context.Context, mediaType models.MediaType, identifier string) (*Result, error) {
	switch mediaType {
	case models.MediaTypeBook:
		return enrichBook(ctx, identifier)
	case models.MediaTypeAudio:
		return enrichAudio(ctx, identifier)
	case models.MediaTypeVideo:
		return enrichVideo(ctx, identifier)
	case models.MediaTypeGame:
		return enrichGame(ctx, identifier)
	case models.MediaTypeThing:
		return nil, fmt.Errorf("THING type requires manual input")
	default:
		return nil, fmt.Errorf("unknown media type: %s", mediaType)
	}
}
