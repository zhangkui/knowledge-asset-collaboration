package user

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

type Credential struct {
	UserID, PasswordHash string
	ChangedAt            time.Time
	MustChange           bool
}
type PasswordService struct {
	mu          sync.RWMutex
	credentials map[string]Credential
	attempts    map[string][]LoginEvent
}

func NewPasswordService() *PasswordService {
	return &PasswordService{credentials: map[string]Credential{}, attempts: map[string][]LoginEvent{}}
}
func digest(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}
func validatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("password must contain at least 8 characters")
	}
	if strings.TrimSpace(password) == "" {
		return errors.New("password required")
	}
	return nil
}
func (p *PasswordService) Set(ctx context.Context, user, password string, mustChange bool) (Credential, error) {
	if ctx == nil {
		return Credential{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Credential{}, err
	}
	if err := validatePassword(password); err != nil {
		return Credential{}, err
	}
	if user == "" {
		return Credential{}, errors.New("user required")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	c := Credential{UserID: user, PasswordHash: digest(password), ChangedAt: time.Now(), MustChange: mustChange}
	p.credentials[user] = c
	return c, nil
}
func (p *PasswordService) Verify(ctx context.Context, user, password, ip, agent string) (bool, error) {
	if ctx == nil {
		return false, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	p.mu.RLock()
	c, ok := p.credentials[user]
	p.mu.RUnlock()
	success := ok && c.PasswordHash == digest(password)
	p.mu.Lock()
	p.attempts[user] = append(p.attempts[user], LoginEvent{UserID: user, IP: ip, UserAgent: agent, Reason: map[bool]string{true: "authenticated", false: "invalid credentials"}[success], Success: success, At: time.Now()})
	p.mu.Unlock()
	return success, nil
}
func (p *PasswordService) Change(ctx context.Context, user, current, next string) (Credential, error) {
	if ctx == nil {
		return Credential{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Credential{}, err
	}
	valid, err := p.Verify(ctx, user, current, "", "password-change")
	if err != nil {
		return Credential{}, err
	}
	if !valid {
		return Credential{}, errors.New("current password invalid")
	}
	return p.Set(ctx, user, next, false)
}
func (p *PasswordService) Credential(ctx context.Context, user string) (Credential, error) {
	if ctx == nil {
		return Credential{}, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return Credential{}, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	c, ok := p.credentials[user]
	if !ok {
		return Credential{}, errors.New("credential not found")
	}
	return c, nil
}
func (p *PasswordService) FailedAttempts(ctx context.Context, user string, since time.Time) int {
	if ctx == nil || ctx.Err() != nil {
		return 0
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	count := 0
	for _, event := range p.attempts[user] {
		if !event.Success && !event.At.Before(since) {
			count++
		}
	}
	return count
}
func (p *PasswordService) LockoutRequired(ctx context.Context, user string, since time.Time, limit int) bool {
	return p.FailedAttempts(ctx, user, since) >= limit
}
func (p *PasswordService) LoginEvents(ctx context.Context, user string) ([]LoginEvent, error) {
	if ctx == nil {
		return nil, errors.New("context is nil")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return append([]LoginEvent(nil), p.attempts[user]...), nil
}
