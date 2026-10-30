package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
	"telegram-bot/inmem"
	"telegram-bot/register"
)

type Subscriptor struct {
	Register      *register.JobRegistry
	Store         *inmem.Store
	Bot           *tele.Bot
	Logger        *zap.Logger
	stream        chan Event
	subscriptions map[string]map[int64]struct{}
	mu            sync.Mutex
	wg            sync.WaitGroup
}

type Event interface {
	ID() string
	String() string
	Source() string
	Image() string
}

func New(bot *tele.Bot, logger *zap.Logger) *Subscriptor {
	subscriptor := &Subscriptor{
		Register:      register.New(),
		Store:         inmem.New(7 * 24 * time.Hour),
		Bot:           bot,
		Logger:        logger,
		stream:        make(chan Event),
		subscriptions: make(map[string]map[int64]struct{}),
	}

	go func() {
		for event := range subscriptor.stream {
			if subscriptor.Store.MarkSeen(event.ID()) {
				continue
			}

			subscriptor.mu.Lock()
			chatIDs := make([]int64, 0, len(subscriptor.subscriptions[event.Source()]))
			for chatID := range subscriptor.subscriptions[event.Source()] {
				chatIDs = append(chatIDs, chatID)
			}
			subscriptor.mu.Unlock()

			for _, chatID := range chatIDs {
				if img := event.Image(); img != "" {
					if _, err := subscriptor.Bot.Send(tele.ChatID(chatID), &tele.Photo{
						File:    tele.FromURL(img),
						Caption: event.String(),
					}); err != nil {
						subscriptor.Logger.Error(
							"failed to send photo",
							zap.Error(err),
							zap.String("chat_id", fmt.Sprintf("%d", chatID)),
						)
					}
				} else {
					if _, err := subscriptor.Bot.Send(tele.ChatID(chatID), event.String(), tele.ModeHTML); err != nil {
						subscriptor.Logger.Error(
							"failed to send event",
							zap.Error(err),
							zap.String("chat_id", fmt.Sprintf("%d", chatID)),
						)
					}
				}
			}
		}
	}()

	return subscriptor
}

func (s *Subscriptor) Start(c tele.Context, method string, job func() ([]Event, error), frequency time.Duration) error {
	chatID := c.Chat().ID

	s.mu.Lock()
	if s.subscriptions[method] == nil {
		s.subscriptions[method] = make(map[int64]struct{})
	}
	s.subscriptions[method][chatID] = struct{}{}
	s.mu.Unlock()

	s.wg.Add(1)
	started := s.Register.StartOnce(method, func(ctx context.Context) {
		defer s.wg.Done()
		ticker := time.NewTicker(frequency)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				messages, err := job()
				if err != nil {
					s.Logger.Error("job failed", zap.Error(err))
					continue
				}

				for _, message := range messages {
					s.stream <- message
				}
			}
		}
	})

	if !started {
		s.wg.Done()
		return nil
	}

	return nil
}

func (s *Subscriptor) Stop(c tele.Context, method string) error {
	s.mu.Lock()
	if s.subscriptions[method] != nil {
		delete(s.subscriptions[method], c.Chat().ID)
	}
	remaining := len(s.subscriptions[method])
	if remaining == 0 {
		delete(s.subscriptions, method)
	}
	s.mu.Unlock()

	if remaining > 0 {
		return nil
	}

	if !s.Register.Stop(method) {
		return fmt.Errorf("%s method could not be stopped", method)
	}

	return nil
}

func (s *Subscriptor) Close() {
	s.mu.Lock()
	for method := range s.subscriptions {
		s.Register.Stop(method)
	}
	s.mu.Unlock()

	s.wg.Wait()
	close(s.stream)
}
