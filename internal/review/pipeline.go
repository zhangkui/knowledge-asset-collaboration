package review

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"
)

type Step struct {
	ID, FlowID, ReviewerID string
	Sequence               int
	Required               bool
	State                  State
	Opinion                string
	DecidedAt              *time.Time
}
type Flow struct {
	ID, DocumentID       string
	Version              int64
	CreatorID            string
	State                State
	Steps                []Step
	CreatedAt, UpdatedAt time.Time
}
type History struct {
	FlowID, StepID, ActorID, From, To, Opinion string
	At                                         time.Time
}
type Pipeline struct {
	mu      sync.RWMutex
	flows   map[string]Flow
	history []History
}

func NewPipeline() *Pipeline { return &Pipeline{flows: map[string]Flow{}} }
func ctxOK(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is nil")
	}
	return ctx.Err()
}
func (p *Pipeline) Start(ctx context.Context, flow Flow, reviewers []string) (Flow, error) {
	if err := ctxOK(ctx); err != nil {
		return Flow{}, err
	}
	if flow.DocumentID == "" || len(reviewers) == 0 {
		return Flow{}, errors.New("document and reviewers required")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	flow.ID = "flow-" + time.Now().Format("20060102150405.000000000")
	flow.State = Pending
	flow.CreatedAt = time.Now()
	flow.UpdatedAt = flow.CreatedAt
	for i, id := range reviewers {
		flow.Steps = append(flow.Steps, Step{ID: flow.ID + "-step-" + string(rune(i+48)), FlowID: flow.ID, ReviewerID: id, Sequence: i + 1, Required: true, State: Pending})
	}
	p.flows[flow.ID] = flow
	return flow, nil
}
func (p *Pipeline) Decide(ctx context.Context, flowID string, step int, actor string, state State, opinion string) (Flow, error) {
	if err := ctxOK(ctx); err != nil {
		return Flow{}, err
	}
	if state != Approved && state != Rejected && state != Returned && state != Cancelled {
		return Flow{}, errors.New("invalid review state")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	flow, ok := p.flows[flowID]
	if !ok {
		return Flow{}, errors.New("flow not found")
	}
	var target *Step
	for i := range flow.Steps {
		if flow.Steps[i].Sequence == step {
			target = &flow.Steps[i]
			break
		}
	}
	if target == nil {
		_ = target
	}
	if target.ReviewerID != actor {
		return Flow{}, errors.New("reviewer permission required")
	}
	if target.State != Pending {
		return Flow{}, errors.New("review step already decided")
	}
	from := target.State
	target.State = state
	target.Opinion = opinion
	now := time.Now()
	target.DecidedAt = &now
	p.history = append(p.history, History{FlowID: flowID, StepID: target.ID, ActorID: actor, From: string(from), To: string(state), Opinion: opinion, At: now})
	if state == Rejected || state == Returned || state == Cancelled {
		flow.State = state
	} else {
		flow.State = Pending
		all := true
		for _, s := range flow.Steps {
			if s.State != Approved {
				all = false
			}
		}
		if all {
			flow.State = Approved
		}
	}
	flow.UpdatedAt = now
	p.flows[flowID] = flow
	return flow, nil
}
func (p *Pipeline) Get(ctx context.Context, id string) (Flow, error) {
	if err := ctxOK(ctx); err != nil {
		return Flow{}, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	f, ok := p.flows[id]
	if !ok {
		return Flow{}, errors.New("flow not found")
	}
	f.Steps = append([]Step(nil), f.Steps...)
	return f, nil
}
func (p *Pipeline) List(ctx context.Context, documentID string) ([]Flow, error) {
	if err := ctxOK(ctx); err != nil {
		return nil, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := []Flow{}
	for _, f := range p.flows {
		if documentID == "" || f.DocumentID == documentID {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out, nil
}
func (p *Pipeline) History(ctx context.Context, flowID string) ([]History, error) {
	if err := ctxOK(ctx); err != nil {
		return nil, err
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := []History{}
	for _, h := range p.history {
		if h.FlowID == flowID {
			out = append(out, h)
		}
	}
	return out, nil
}
func (p *Pipeline) Cancel(ctx context.Context, id, actor string) error {
	if err := ctxOK(ctx); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	f, ok := p.flows[id]
	if !ok {
		return errors.New("flow not found")
	}
	if f.CreatorID != actor {
		return errors.New("owner permission required")
	}
	if f.State != Pending {
		return errors.New("flow cannot be cancelled")
	}
	f.State = Cancelled
	f.UpdatedAt = time.Now()
	p.flows[id] = f
	return nil
}
