package inmem

import (
	"sync"

	"telegram-bot/scraper/belasartes"
)

type Store struct {
	mu             sync.Mutex
	seenBelasArtes map[int64]map[belasartes.Event]struct{}
}

func New() *Store {
	return &Store{
		seenBelasArtes: make(map[int64]map[belasartes.Event]struct{}),
	}
}

func (s *Store) MarkSeenBelasArtes(chatID int64, events []belasartes.Event) []belasartes.Event {
	s.mu.Lock()
	defer s.mu.Unlock()

	seen := s.seenBelasArtes[chatID]
	if seen == nil {
		seen = make(map[belasartes.Event]struct{})
		s.seenBelasArtes[chatID] = seen
	}

	toSend := make([]belasartes.Event, 0)
	for _, event := range events {
		if _, ok := seen[event]; ok {
			continue
		}
		seen[event] = struct{}{}
		toSend = append(toSend, event)
	}
	return toSend
}
