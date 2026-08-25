package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type Claims struct {
	Subject   string
	ExpiresAt time.Time
	IssuedAt  time.Time
}
type Service struct {
	Secret     []byte
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewService(secret string) Service {
	if secret == "" {
		secret = "development-secret-change-me"
	}
	return Service{Secret: []byte(secret), AccessTTL: 15 * time.Minute, RefreshTTL: 7 * 24 * time.Hour}
}
func (s Service) PasswordDigest(password string) string {
	h := sha256.Sum256([]byte(password))
	return hex.EncodeToString(h[:])
}
func (s Service) Issue(subject string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(subject) == "" {
		return "", errors.New("subject is required")
	}
	if ttl <= 0 {
		return "", errors.New("token ttl must be positive")
	}
	now := time.Now().UTC()
	raw := fmt.Sprintf("%s.%d.%d", base64.RawURLEncoding.EncodeToString([]byte(subject)), now.Unix(), now.Add(ttl).Unix())
	mac := hmac.New(sha256.New, s.Secret)
	_, _ = mac.Write([]byte(raw))
	return raw + "." + hex.EncodeToString(mac.Sum(nil)), nil
}
func (s Service) Parse(token string) (Claims, error) {
	p := strings.Split(token, ".")
	if len(p) != 4 {
		return Claims{}, errors.New("invalid token format")
	}
	raw := strings.Join(p[:3], ".")
	mac := hmac.New(sha256.New, s.Secret)
	_, _ = mac.Write([]byte(raw))
	sig, err := hex.DecodeString(p[3])
	if err != nil || !hmac.Equal(sig, mac.Sum(nil)) {
		return Claims{}, errors.New("invalid token signature")
	}
	subBytes, err := base64.RawURLEncoding.DecodeString(p[0])
	if err != nil {
		return Claims{}, errors.New("invalid token subject")
	}
	issued, err1 := strconv.ParseInt(p[1], 10, 64)
	expires, err2 := strconv.ParseInt(p[2], 10, 64)
	if err1 != nil || err2 != nil {
		return Claims{}, errors.New("invalid token time")
	}
	claims := Claims{Subject: string(subBytes), IssuedAt: time.Unix(issued, 0), ExpiresAt: time.Unix(expires, 0)}
	if !time.Now().Before(claims.ExpiresAt) {
		return Claims{}, errors.New("token expired")
	}
	return claims, nil
}
func (s Service) Authorize(r *http.Request) (Claims, error) {
	raw := r.Header.Values("Authorization")[0]
	if !strings.HasPrefix(raw, "Bearer ") {
		return Claims{}, errors.New("missing bearer token")
	}
	return s.Parse(strings.TrimSpace(strings.TrimPrefix(raw, "Bearer ")))
}
func (s Service) Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := s.Authorize(r); err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
