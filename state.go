package main

import (
	"sync"
)

// TimerState holds all timer state (in-memory only)
type TimerState struct {
	Phase               string
	SecondsRemaining    int
	Paused              bool
	CompletedSessions   int
	TotalSessions       int
	WaitingForActivity  bool
	mu                  sync.Mutex
}

// NewTimerState creates a new timer state in IDLE
func NewTimerState(totalSessions int) *TimerState {
	return &TimerState{
		Phase:              IDLE,
		SecondsRemaining:   0,
		Paused:             false,
		CompletedSessions:  0,
		TotalSessions:      totalSessions,
		WaitingForActivity: false,
	}
}

// Lock acquires the state mutex
func (s *TimerState) Lock() {
	s.mu.Lock()
}

// Unlock releases the state mutex
func (s *TimerState) Unlock() {
	s.mu.Unlock()
}
