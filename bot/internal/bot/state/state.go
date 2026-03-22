package state

import "sync"

type PendingKind string

const (
	PendingNone       PendingKind = ""
	PendingDoneReport PendingKind = "done_report"
	PendingDoneFlow   PendingKind = "done_flow"
	PendingRemind     PendingKind = "remind_create"
	PendingPlanAdd    PendingKind = "plan_add"
	PendingSettings   PendingKind = "settings"
)

type DoneExercise struct {
	Name string
	Plan string
}

type DoneAnswer struct {
	Name   string
	Plan   string
	Actual string
}

type DoneFlowSession struct {
	WorkoutTitle    string
	Exercises       []DoneExercise
	Answers         []DoneAnswer
	CurrentIndex    int
	PromptMessageID int
	SourceReminder  *int64
}

type PlanDraftExercise struct {
	Name string
	Plan string
}

type PlanDraftStep string

const (
	PlanDraftSelectDays PlanDraftStep = "select_days"
	PlanDraftTitle      PlanDraftStep = "title"
	PlanDraftExercises  PlanDraftStep = "exercises"
)

type PlanDraftSession struct {
	Step      PlanDraftStep
	Days      []int
	Title     string
	Exercises []PlanDraftExercise
	PromptMsg int
}

type Pending struct {
	Kind       PendingKind
	ReminderID *int64
	DoneFlow   *DoneFlowSession
	PlanDraft  *PlanDraftSession
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
