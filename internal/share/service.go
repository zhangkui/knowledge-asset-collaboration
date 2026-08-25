package share

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
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
