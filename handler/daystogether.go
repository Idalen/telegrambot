package handler

import (
	"fmt"
	"time"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

const _message = "Hoje os momorzinhos fazem %d dias juntos! 🥰❤️"

func (h *Handler) DaysTogether(c tele.Context) error {
	h.Logger.Info("command actioned", zap.String("command", "/daystogether"))
	days := daysTogether()
	return c.Send(fmt.Sprintf(_message, days))
}

func (h *Handler) DaysTogetherJob(chatID int64) {
	days := daysTogether()
	h.Bot.Send(tele.ChatID(chatID), fmt.Sprintf(_message, days))
}

func daysTogether() int {
	dateWeMet := time.Date(2023, 8, 6, 0, 0, 0, 0, time.Local)
	now := time.Now()
	return int(now.Sub(dateWeMet).Hours() / 24)
}
