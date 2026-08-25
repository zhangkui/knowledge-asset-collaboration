package user

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

type Profile struct {
	UserID, DisplayName, Phone, AvatarURL, Bio string
	UpdatedAt                                  time.Time
}
type Department struct {
	ID, Name, ParentID, ManagerID string
	Enabled                       bool
	CreatedAt, UpdatedAt          time.Time
}
type Team struct {
	ID, Name, OwnerID    string
	MemberIDs            []string
	CreatedAt, UpdatedAt time.Time
}
type Role struct {
	ID, Name    string
	Permissions []string
	BuiltIn     bool
}
type LoginEvent struct {
	UserID, IP, UserAgent, Reason string
	Success                       bool
	At                            time.Time
}
type Directory struct {
	mu          sync.RWMutex
	users       map[string]User
	profiles    map[string]Profile
	departments map[string]Department
	teams       map[string]Team
	roles       map[string]Role
	assignments map[string]map[string]bool
	logins      []LoginEvent
}

func NewDirectory() *Directory {
	return &Directory{users: map[string]User{}, profiles: map[string]Profile{}, departments: map[string]Department{}, teams: map[string]Team{}, roles: map[string]Role{}, assignments: map[string]map[string]bool{}}
}
func check(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	return ctx.Err()
}
func (d *Directory) Register(ctx context.Context, u User) (User, error) {
	if err := check(ctx); err != nil {
		return User{}, err
	}
	if strings.TrimSpace(u.Email) == "" {
		return User{}, errors.New("email required")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for _, existing := range d.users {
		if strings.EqualFold(existing.Email, u.Email) {
			return User{}, errors.New("email already exists")
		}
	}
	created := time.Now()
	if u.ID == "" {
		u.ID = "usr-" + time.Now().Format("20060102150405.000000000")
	}
	u.Enabled = true
	u.CreatedAt = created
	d.users[u.ID] = u
	d.profiles[u.ID] = Profile{UserID: u.ID, DisplayName: u.Name, UpdatedAt: created}
	return u, nil
}
func (d *Directory) UpdateProfile(ctx context.Context, id string, p Profile) (Profile, error) {
	if err := check(ctx); err != nil {
		return Profile{}, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.users[id]; !ok {
		return Profile{}, errors.New("user not found")
	}
	p.UserID = id
	p.UpdatedAt = time.Now()
	d.profiles[id] = p
	return p, nil
}
func (d *Directory) GetProfile(ctx context.Context, id string) (Profile, error) {
	if err := check(ctx); err != nil {
		return Profile{}, err
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	p, ok := d.profiles[id]
	if !ok {
		return Profile{}, errors.New("profile not found")
	}
	return p, nil
}
func (d *Directory) SetEnabled(ctx context.Context, id string, enabled bool) error {
	if err := check(ctx); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	u, ok := d.users[id]
	if !ok {
		return errors.New("user not found")
	}
	u.Enabled = enabled
	d.users[id] = u
	return nil
}
func (d *Directory) ListUsers(ctx context.Context, departmentID string) ([]User, error) {
	if err := check(ctx); err != nil {
		return nil, err
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := []User{}
	for _, u := range d.users {
		if departmentID == "" || u.DepartmentID == departmentID {
			out = append(out, u)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Email < out[j].Email })
	return out, nil
}
func (d *Directory) CreateDepartment(ctx context.Context, name, parent, manager string) (Department, error) {
	if err := check(ctx); err != nil {
		return Department{}, err
	}
	if strings.TrimSpace(name) == "" {
		return Department{}, errors.New("department name required")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if parent != "" {
		if _, ok := d.departments[parent]; !ok {
			return Department{}, errors.New("parent department not found")
		}
	}
	dep := Department{ID: "dep-" + time.Now().Format("20060102150405.000000000"), Name: strings.TrimSpace(name), ParentID: parent, ManagerID: manager, Enabled: true, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	d.departments[dep.ID] = dep
	return dep, nil
}
func (d *Directory) UpdateDepartment(ctx context.Context, id, name, manager string, enabled bool) (Department, error) {
	if err := check(ctx); err != nil {
		return Department{}, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	dep, ok := d.departments[id]
	if !ok {
		return Department{}, errors.New("department not found")
	}
	if strings.TrimSpace(name) != "" {
		dep.Name = strings.TrimSpace(name)
	}
	dep.ManagerID = manager
	dep.Enabled = enabled
	dep.UpdatedAt = time.Now()
	d.departments[id] = dep
	return dep, nil
}
func (d *Directory) ListDepartments(ctx context.Context) ([]Department, error) {
	if err := check(ctx); err != nil {
		return nil, err
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := make([]Department, 0, len(d.departments))
	for _, dep := range d.departments {
		out = append(out, dep)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
func (d *Directory) CreateTeam(ctx context.Context, name, owner string) (Team, error) {
	if err := check(ctx); err != nil {
		return Team{}, err
	}
	if strings.TrimSpace(name) == "" {
		return Team{}, errors.New("team name required")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.users[owner]; !ok {
		return Team{}, errors.New("owner not found")
	}
	team := Team{ID: "team-" + time.Now().Format("20060102150405.000000000"), Name: name, OwnerID: owner, MemberIDs: []string{owner}, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	d.teams[team.ID] = team
	return team, nil
}
func (d *Directory) AddTeamMember(ctx context.Context, teamID, userID string) (Team, error) {
	if err := check(ctx); err != nil {
		return Team{}, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	team, ok := d.teams[teamID]
	if !ok {
		return Team{}, errors.New("team not found")
	}
	if _, ok := d.users[userID]; !ok {
		return Team{}, errors.New("user not found")
	}
	for _, id := range team.MemberIDs {
		if id == userID {
			return team, nil
		}
	}
	team.MemberIDs = append(team.MemberIDs, userID)
	team.UpdatedAt = time.Now()
	d.teams[teamID] = team
	return team, nil
}
func (d *Directory) RemoveTeamMember(ctx context.Context, teamID, userID string) (Team, error) {
	if err := check(ctx); err != nil {
		return Team{}, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	team, ok := d.teams[teamID]
	if !ok {
		return Team{}, errors.New("team not found")
	}
	if team.OwnerID == userID {
		return Team{}, errors.New("team owner cannot leave")
	}
	next := team.MemberIDs[:0]
	for _, id := range team.MemberIDs {
		if id != userID {
			next = append(next, id)
		}
	}
	team.MemberIDs = next
	team.UpdatedAt = time.Now()
	d.teams[teamID] = team
	return team, nil
}
func (d *Directory) DefineRole(ctx context.Context, name string, permissions []string, builtIn bool) (Role, error) {
	if err := check(ctx); err != nil {
		return Role{}, err
	}
	if strings.TrimSpace(name) == "" {
		return Role{}, errors.New("role name required")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	copyPerms := append([]string(nil), permissions...)
	role := Role{ID: "role-" + time.Now().Format("20060102150405.000000000"), Name: name, Permissions: copyPerms, BuiltIn: builtIn}
	d.roles[role.ID] = role
	return role, nil
}
func (d *Directory) AssignRole(ctx context.Context, userID, roleID string) error {
	if err := check(ctx); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if _, ok := d.users[userID]; !ok {
		return errors.New("user not found")
	}
	if _, ok := d.roles[roleID]; !ok {
		return errors.New("role not found")
	}
	if d.assignments[userID] == nil {
		d.assignments[userID] = map[string]bool{}
	}
	d.assignments[userID][roleID] = true
	return nil
}
func (d *Directory) RolesForUser(ctx context.Context, userID string) ([]Role, error) {
	if err := check(ctx); err != nil {
		return nil, err
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	ids := d.assignments[userID]
	out := []Role{}
	for id := range ids {
		if role, ok := d.roles[id]; ok {
			out = append(out, role)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}
func (d *Directory) RecordLogin(ctx context.Context, event LoginEvent) error {
	if err := check(ctx); err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	event.At = time.Now()
	d.logins = append(d.logins, event)
	return nil
}
func (d *Directory) LoginHistory(ctx context.Context, userID string, limit int) ([]LoginEvent, error) {
	if err := check(ctx); err != nil {
		return nil, err
	}
	d.mu.RLock()
	defer d.mu.RUnlock()
	out := []LoginEvent{}
	for i := len(d.logins) - 1; i >= 0; i-- {
		if userID == "" || d.logins[i].UserID == userID {
			out = append(out, d.logins[i])
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}
