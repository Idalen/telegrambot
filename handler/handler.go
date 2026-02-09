package handler

import (
	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

type Handler struct {
	Bot    *tele.Bot
	Logger *zap.Logger
}

func (h *Handler) Start(c tele.Context) error {
	h.Logger.Info("command actioned", zap.String("command", "/ping"))
	return c.Send("Oi momorzinhos! ♡ ")
}
