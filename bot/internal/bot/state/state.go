package state

import "sync"

type PendingKind string

const (
	PendingNone       PendingKind = ""
	PendingDoneReport PendingKind = "done_report"
	PendingRemind     PendingKind = "remind_create"
	PendingSettings   PendingKind = "settings"
)

type Pending struct {
	Kind       PendingKind
	ReminderID *int64
}

type Store struct {
	mu sync.Mutex
	m  map[int64]Pending // tg_id -> pending
}

func New() *Store {
	return &Store{m: make(map[int64]Pending)}
}

func (s *Store) Set(tgID int64, p Pending) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[tgID] = p
}

func (s *Store) Get(tgID int64) (Pending, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.m[tgID]
	return p, ok
}

func (s *Store) Clear(tgID int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, tgID)
}

