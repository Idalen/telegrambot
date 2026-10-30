package inmem

import (
	"sync"
	"time"
)

type Store struct {
	mu   sync.Mutex
	seen map[string]time.Time
	ttl  time.Duration
}

func New(ttl time.Duration) *Store {
	return &Store{
		seen: make(map[string]time.Time),
		ttl:  ttl,
	}
}

func (s *Store) MarkSeen(eventID string) (alreadySeen bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for id, ts := range s.seen {
		if now.Sub(ts) >= s.ttl {
			delete(s.seen, id)
		}
	}

	if _, ok := s.seen[eventID]; ok {
		return true
	}

	s.seen[eventID] = now
	return false
}
