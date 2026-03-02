package models

import (
	"time"

	"github.com/google/uuid"
)

type Role string

const (
	RoleUser    Role = "USER"
	RoleManager Role = "MANAGER"
)

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	DisplayName  string    `json:"displayName"`
	PasswordHash string    `json:"-"`
	SSOProvider  string    `json:"ssoProvider,omitempty"` // e.g. "google"
	SSOSubject   string    `json:"ssoSubject,omitempty"`  // provider's user ID
	CreatedAt    time.Time `json:"createdAt"`
}

type NodeClaim struct {
	UserID    uuid.UUID `json:"userId"`
	NodeID    string    `json:"nodeId"` // node public key fingerprint
	Role      Role      `json:"role"`
	GrantedBy uuid.UUID `json:"grantedBy"`
	GrantedAt time.Time `json:"grantedAt"`
}

// JWTClaim is the node-scoped claim embedded in issued JWTs.
type JWTClaim struct {
	Node string `json:"node"`
	Role Role   `json:"role"`
}
