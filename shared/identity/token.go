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
	UserID string            `json:"uid"`
	Claims []models.JWTClaim `json:"claims"`
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
