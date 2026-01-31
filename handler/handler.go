package handler

import (
	"telegram-bot/register"

	tele "gopkg.in/telebot.v3"
)

type Handler struct {
	Bot *tele.Bot
	Register *register.JobRegistry
}

func (h *Handler) Start(c tele.Context) error {
	return c.Send("Oi momorzinhos! ♡ ")
}


