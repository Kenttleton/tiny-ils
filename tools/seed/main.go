// seed populates one node's curios-manager with test data via gRPC.
// It tries EnrichMetadata first and falls back to hardcoded data on failure.
//
// Usage:
//
//	CURIOS_GRPC=localhost:50151 go run ./tools/seed
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	pb "tiny-ils/gen/curiospb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type seedItem struct {
	mediaType  string
	formatType string
	identifier string // passed to EnrichMetadata
	fileRef    string // for digital assets: "blank.epub" or "blank.pdf"
	fallback   pb.CreateCurioRequest
}

var seeds = []seedItem{
	{
		mediaType:  "BOOK",
		formatType: "PHYSICAL",
		identifier: "9780441013593",
		fallback: pb.CreateCurioRequest{
			Title:       "Dune",
			Description: "A science fiction novel by Frank Herbert set in the far future on the desert planet Arrakis.",
			Tags:        []string{"sci-fi", "classic", "epic"},
		},
	},
	{
		mediaType:  "BOOK",
		formatType: "BOTH",
		identifier: "9780553293357",
		fileRef:    "blank.epub",
		fallback: pb.CreateCurioRequest{
			Title:       "Foundation",
			Description: "Isaac Asimov's epic saga of the fall and rise of a Galactic Empire.",
			Tags:        []string{"sci-fi", "classic"},
		},
	},
	{
		mediaType:  "BOOK",
		formatType: "DIGITAL",
		identifier: "9780141439518",
		fileRef:    "blank.epub",
		fallback: pb.CreateCurioRequest{
			Title:       "Pride and Prejudice",
			Description: "Jane Austen's beloved novel of manners, marriage, and morality in early 19th-century England.",
			Tags:        []string{"classic", "romance", "fiction"},
		},
	},
	{
		mediaType:  "BOOK",
		formatType: "BOTH",
		identifier: "9780553418026",
		fileRef:    "blank.epub",
		fallback: pb.CreateCurioRequest{
			Title:       "The Martian",
			Description: "An astronaut stranded on Mars must use his ingenuity to survive until rescue.",
			Tags:        []string{"sci-fi", "survival", "humor"},
		},
	},
	{
		mediaType:  "AUDIO",
		formatType: "PHYSICAL",
		identifier: "The Dark Side of the Moon Pink Floyd",
		fallback: pb.CreateCurioRequest{
			Title:       "The Dark Side of the Moon",
			Description: "The classic 1973 album by Pink Floyd.",
			Tags:        []string{"rock", "progressive rock", "classic"},
		},
	},
	{
		mediaType:  "GAME",
		formatType: "PHYSICAL",
		identifier: "Minecraft",
		fallback: pb.CreateCurioRequest{
			Title:       "Minecraft",
			Description: "A sandbox video game developed by Mojang Studios.",
			Tags:        []string{"sandbox", "survival", "building"},
		},
	},
}

func main() {
	addr := os.Getenv("CURIOS_GRPC")
	if addr == "" {
		addr = "localhost:50151"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("connect %s: %v", addr, err)
	}
	defer conn.Close()

	client := pb.NewCuriosManagerClient(conn)

	fmt.Printf("seeding %s\n", addr)
	ok, skipped := 0, 0
	for _, s := range seeds {
		if err := seed(ctx, client, s); err != nil {
			log.Printf("SKIP %q: %v", s.fallback.Title, err)
			skipped++
		} else {
			ok++
		}
	}
	fmt.Printf("done: %d seeded, %d skipped\n", ok, skipped)
}

func seed(ctx context.Context, client pb.CuriosManagerClient, s seedItem) error {
	req := s.fallback
	req.MediaType = s.mediaType
	req.FormatType = s.formatType

	// Attempt metadata enrichment; fall back silently on error.
	meta, err := client.EnrichMetadata(ctx, &pb.EnrichRequest{
		MediaType:  s.mediaType,
		Identifier: s.identifier,
	})
	if err != nil {
		fmt.Printf("  enrich %q: %v — using fallback\n", req.Title, err)
	} else {
		if meta.Title != "" {
			req.Title = meta.Title
		}
		if meta.Description != "" {
			req.Description = meta.Description
		}
		if len(meta.Tags) > 0 {
			req.Tags = meta.Tags
		}
	}

	curio, err := client.CreateCurio(ctx, &req)
	if err != nil {
		return fmt.Errorf("CreateCurio: %w", err)
	}
	fmt.Printf("  + curio  %q  [%s/%s]  %s\n", curio.Title, curio.MediaType, curio.FormatType, curio.Id[:8])

	// Physical copies — 2 per item.
	if s.formatType == "PHYSICAL" || s.formatType == "BOTH" {
		for _, loc := range []string{"Shelf A1", "Shelf A2"} {
			_, err := client.CreateCopy(ctx, &pb.CreateCopyRequest{
				CurioId:   curio.Id,
				Condition: "GOOD",
				Location:  loc,
			})
			if err != nil {
				log.Printf("    copy at %s: %v", loc, err)
			} else {
				fmt.Printf("    + copy   %s\n", loc)
			}
		}
	}

	// Digital asset.
	if s.formatType == "DIGITAL" || s.formatType == "BOTH" {
		fileRef := s.fileRef
		format := "EPUB"
		if fileRef == "blank.pdf" {
			format = "PDF"
		}
		_, err := client.CreateDigitalAsset(ctx, &pb.CreateDigitalAssetRequest{
			CurioId:       curio.Id,
			Format:        format,
			FileRef:       fileRef,
			MaxConcurrent: 5,
		})
		if err != nil {
			log.Printf("    digital asset: %v", err)
		} else {
			fmt.Printf("    + digital %s (%s)\n", format, fileRef)
		}
	}

	return nil
}
