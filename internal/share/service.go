package share

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// Sentinel errors for share resolution. Callers should inspect them with
// errors.Is so the failure cause survives wrapping (no %v that breaks the
// chain).
var (
	// ErrShareNotFound is returned when a share token does not exist (for
	// example an empty link or an unknown token).
	ErrShareNotFound = errors.New("share not found")
	// ErrShareExpired is returned when a share token exists but is past its
	// expiry.
	ErrShareExpired = errors.New("share expired")
	// ErrShareRevoked is returned when a share token has been revoked.
	ErrShareRevoked = errors.New("share revoked")
	// ErrInvalidShareLink is returned when a share link is structurally
	// invalid (missing token or document).
	ErrInvalidShareLink = errors.New("invalid share link")
)

type Link struct {
	ID, DocumentID, Permission, Token string
	ExpiresAt                         time.Time
}

func Create(ctx context.Context, document, permission string, ttl time.Duration) (Link, error) {
	if err := ctx.Err(); err != nil {
		return Link{}, err
	}
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return Link{}, err
	}
	return Link{ID: fmt.Sprintf("share-%d", time.Now().UnixNano()), DocumentID: document, Permission: permission, Token: hex.EncodeToString(buf), ExpiresAt: time.Now().Add(ttl)}, nil
}
func Valid(l Link, now time.Time) bool { return l.Token != "" && now.Before(l.ExpiresAt) }
