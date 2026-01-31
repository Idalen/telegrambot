package handler

import (
	"context"
	"fmt"
	"time"

	tele "gopkg.in/telebot.v3"
)

const (
	_message = "Hoje os momorzinhos fazem %d dias juntos! 🥰❤️"
	_method = "dayssincewemet"
)

func (h *Handler) DaysSinceWeMet(c tele.Context) error {
	chatID := c.Chat().ID

	started := h.Register.StartOnce(chatID, _method, func(ctx context.Context) {
		h.daysSinceWeMetJob(ctx, c.Chat().ID)
	})

	if !started {
		return c.Send(fmt.Sprintf("Already running: %s", _method))
	}
	return c.Send(fmt.Sprintf("Started %s", _method))
}

func (h Handler) StopDaysSinceWeMet(c tele.Context) error { 
	if !h.Register.Stop(c.Chat().ID, _method) {
		return fmt.Errorf("%s method could not be stopped", _method)
	}

	return nil
}

func (h *Handler) daysSinceWeMetJob(ctx context.Context, chatID int64) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	dateWeMet := time.Date(2023, 8, 6, 0, 0, 0, 0, time.Local)

	sendDays := func() {
		now := time.Now()
		daysSinceWeMet := int(now.Sub(dateWeMet).Hours() / 24)
		h.Bot.Send(tele.ChatID(chatID), fmt.Sprintf(_message, daysSinceWeMet))
	}

	sendDays()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sendDays()
		}
	}
}
