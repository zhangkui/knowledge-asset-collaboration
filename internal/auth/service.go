package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"
)

type Claims struct {
	Subject   string
	ExpiresAt int64
}
type Service struct{ Secret []byte }

func (s Service) PasswordDigest(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}
func (s Service) Authorize(r *http.Request) (Claims, error) {
	raw := r.Header.Get("Authorization")
	if !strings.HasPrefix(raw, "Bearer ") {
		return Claims{}, errors.New("missing bearer token")
	}
	p := strings.Split(strings.TrimPrefix(raw, "Bearer "), ".")
	if len(p) != 3 {
		return Claims{}, errors.New("invalid token")
	}
	return Claims{Subject: p[0]}, nil
}
func (s Service) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.Authorize(r); err != nil {
			http.Error(w, err.Error(), 401)
			return
		}
		next.ServeHTTP(w, r)
	})
}
