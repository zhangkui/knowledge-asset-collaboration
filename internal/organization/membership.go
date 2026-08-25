package organization

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

type Member struct {
	UserID, ScopeID, ScopeType, Role string
	JoinedAt                         time.Time
	Enabled                          bool
}
type ManagedTeam struct {
	ID, Name, OwnerID string
	Members           map[string]bool
	CreatedAt         time.Time
}
type WorkspacePolicy struct {
	WorkspaceID                                 string
	AllowGuest, AllowPublicShare, RequireReview bool
	MaxAttachmentBytes                          int64
	UpdatedAt                                   time.Time
}
type Service struct {
	mu          sync.RWMutex
	members     map[string]Member
	teams       map[string]ManagedTeam
	policies    map[string]WorkspacePolicy
	departments map[string]string
	ancestors   map[string][]string
}

func NewService() *Service {
	return &Service{members: map[string]Member{}, teams: map[string]ManagedTeam{}, policies: map[string]WorkspacePolicy{}, departments: map[string]string{}, ancestors: map[string][]string{}}
}
func validContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	return ctx.Err()
}
func key(user, scope string) string { return user + "@" + scope }
func (s *Service) AddMember(ctx context.Context, user, scope, scopeType, role string) (Member, error) {
	if err := validContext(ctx); err != nil {
		return Member{}, err
	}
	if strings.TrimSpace(user) == "" || strings.TrimSpace(scope) == "" {
		return Member{}, errors.New("member scope required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m := Member{UserID: user, ScopeID: scope, ScopeType: scopeType, Role: role, JoinedAt: time.Now(), Enabled: true}
	s.members[key(user, scope)] = m
	return m, nil
}
func (s *Service) RemoveMember(ctx context.Context, user, scope string) error {
	if err := validContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.members[key(user, scope)]; !ok {
		return errors.New("member not found")
	}
	delete(s.members, key(user, scope))
	return nil
}
func (s *Service) SetMemberRole(ctx context.Context, user, scope, role string) (Member, error) {
	if err := validContext(ctx); err != nil {
		return Member{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.members[key(user, scope)]
	if !ok {
		return Member{}, errors.New("member not found")
	}
	m.Role = role
	m.Enabled = true
	s.members[key(user, scope)] = m
	return m, nil
}
func (s *Service) ListMembers(ctx context.Context, scope string) ([]Member, error) {
	if err := validContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []Member{}
	for _, m := range s.members {
		if scope == "" || m.ScopeID == scope {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UserID < out[j].UserID })
	return out, nil
}
func (s *Service) CreateTeam(ctx context.Context, id, name, owner string) (ManagedTeam, error) {
	if err := validContext(ctx); err != nil {
		return ManagedTeam{}, err
	}
	if name == "" || id == "" {
		return ManagedTeam{}, errors.New("team id and name required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.teams[id]; ok {
		return ManagedTeam{}, errors.New("team already exists")
	}
	t := ManagedTeam{ID: id, Name: name, OwnerID: owner, Members: map[string]bool{owner: true}, CreatedAt: time.Now()}
	s.teams[id] = t
	return t, nil
}
func (s *Service) AddToTeam(ctx context.Context, team, user string) (ManagedTeam, error) {
	if err := validContext(ctx); err != nil {
		return ManagedTeam{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.teams[team]
	if !ok {
		return ManagedTeam{}, errors.New("team not found")
	}
	t.Members[user] = true
	s.teams[team] = t
	return t, nil
}
func (s *Service) RemoveFromTeam(ctx context.Context, team, user string) (ManagedTeam, error) {
	if err := validContext(ctx); err != nil {
		return ManagedTeam{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.teams[team]
	if !ok {
		return ManagedTeam{}, errors.New("team not found")
	}
	if user == t.OwnerID {
		return ManagedTeam{}, errors.New("owner cannot be removed")
	}
	delete(t.Members, user)
	s.teams[team] = t
	return t, nil
}
func (s *Service) TeamMembers(ctx context.Context, team string) ([]string, error) {
	if err := validContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.teams[team]
	if !ok {
		return nil, errors.New("team not found")
	}
	out := make([]string, 0, len(t.Members))
	for id := range t.Members {
		out = append(out, id)
	}
	sort.Strings(out)
	return out, nil
}
func (s *Service) SetDepartmentParent(ctx context.Context, department, parent string) error {
	if err := validContext(ctx); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if department == parent {
		return errors.New("department cycle")
	}
	for current := parent; current != ""; current = s.departments[current] {
		if current == department {
			return errors.New("department cycle")
		}
	}
	s.departments[department] = parent
	s.ancestors[department] = append([]string(nil), s.ancestors[parent]...)
	if parent != "" {
		s.ancestors[department] = append(s.ancestors[department], parent)
	}
	return nil
}
func (s *Service) DepartmentAncestors(ctx context.Context, department string) ([]string, error) {
	if err := validContext(ctx); err != nil {
		return nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	return append([]string(nil), s.ancestors[department]...), nil
}
func (s *Service) SetPolicy(ctx context.Context, policy WorkspacePolicy) (WorkspacePolicy, error) {
	if err := validContext(ctx); err != nil {
		return WorkspacePolicy{}, err
	}
	if policy.WorkspaceID == "" {
		return WorkspacePolicy{}, errors.New("workspace required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	policy.UpdatedAt = time.Now()
	if policy.MaxAttachmentBytes <= 0 {
		policy.MaxAttachmentBytes = 50 << 20
	}
	s.policies[policy.WorkspaceID] = policy
	return policy, nil
}
func (s *Service) Policy(ctx context.Context, workspace string) (WorkspacePolicy, error) {
	if err := validContext(ctx); err != nil {
		return WorkspacePolicy{}, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.policies[workspace]
	if !ok {
		return WorkspacePolicy{}, errors.New("policy not found")
	}
	return p, nil
}
