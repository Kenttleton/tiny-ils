package identity

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
	"tiny-ils/shared/models"

	"github.com/golang-jwt/jwt/v5"
)

// Fingerprint returns a short base64url string identifying a public key.
func Fingerprint(pub ed25519.PublicKey) string {
	h := sha256.Sum256(pub)
	return base64.RawURLEncoding.EncodeToString(h[:16])
}

type NodeClaims struct {
	jwt.RegisteredClaims
	UserID   string            `json:"uid"`
	Claims   []models.JWTClaim `json:"claims"`
	HomeNode string            `json:"home_node,omitempty"` // set for cross-node users; empty for local users
}

// IssueToken signs a JWT for the given user using the node's private key.
func IssueToken(userID, nodeFingerprint string, claims []models.JWTClaim, priv ed25519.PrivateKey, ttl time.Duration) (string, error) {
	now := time.Now()
	c := NodeClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    nodeFingerprint,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID: userID,
		Claims: claims,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, c)
	return token.SignedString(priv)
}

// IssueTokenWithHomeNode signs a local session JWT that carries home_node metadata.
// Used when a cross-node user is granted a local session on a visiting node:
// the JWT is issued by the visiting node but records the user's actual home node.
func IssueTokenWithHomeNode(userID, nodeFingerprint, homeNode string, claims []models.JWTClaim, priv ed25519.PrivateKey, ttl time.Duration) (string, error) {
	now := time.Now()
	c := NodeClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    nodeFingerprint,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID:   userID,
		Claims:   claims,
		HomeNode: homeNode,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, c)
	return token.SignedString(priv)
}

// IssueTokenForAudience signs a short-lived, audience-scoped JWT for cross-node use.
// The token is signed by the home node (iss = nodeFingerprint) and is only valid
// for consumption by audienceNode. No role claims are included — the receiving node
// determines the user's role from its own node_claims table.
func IssueTokenForAudience(userID, nodeFingerprint, audienceNode string, priv ed25519.PrivateKey, ttl time.Duration) (string, error) {
	now := time.Now()
	c := NodeClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    nodeFingerprint,
			Audience:  jwt.ClaimStrings{audienceNode},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
		UserID: userID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, c)
	return token.SignedString(priv)
}

// VerifyToken verifies a JWT using the provided public key and returns its claims.
func VerifyToken(tokenStr string, pub ed25519.PublicKey) (*NodeClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &NodeClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodEd25519); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return pub, nil
	})
	if err != nil {
		return nil, fmt.Errorf("verify token: %w", err)
	}
	c, ok := token.Claims.(*NodeClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}
	return c, nil
}
