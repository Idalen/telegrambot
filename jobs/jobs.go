package jobs

import (
	"sync"
	"time"

	"telegram-bot/inmem"
	"telegram-bot/message"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

type Event interface {
	ID() string
	Message() message.Message
	Source() string
}

type Sender interface {
	Send(tele.Recipient, interface{}, ...interface{}) (*tele.Message, error)
}

// Broadcaster forwards events to all subscribed chats.
type Broadcaster struct {
	Bot    Sender
	Logger *zap.Logger
	Store  *inmem.Store
	chats  map[int64]struct{}
	mu     sync.Mutex
}

func New(bot Sender, logger *zap.Logger) *Broadcaster {
	return &Broadcaster{Bot: bot, Logger: logger, Store: inmem.New(7 * 24 * time.Hour), chats: make(map[int64]struct{})}
}

func (s *Broadcaster) Subscribe(chatID int64) {
	s.mu.Lock()
	s.chats[chatID] = struct{}{}
	s.mu.Unlock()
}

func (s *Broadcaster) Run(events <-chan Event, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			s.mu.Lock()
			chatIDs := make([]int64, 0, len(s.chats))
			for chatID := range s.chats {
				chatIDs = append(chatIDs, chatID)
			}
			s.mu.Unlock()
			if len(chatIDs) == 0 || s.Store.MarkSeen(event.ID()) {
				continue
			}

			msg, options := mapMessage(event.Message())
			for _, chatID := range chatIDs {
				if _, err := s.Bot.Send(tele.ChatID(chatID), msg, options); err != nil {
					s.Logger.Error("failed to send event", zap.Error(err), zap.Int64("chat_id", chatID))
				}
			}
		}
	}
}
