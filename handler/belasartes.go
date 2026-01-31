package handler

import (
	"fmt"
	"strings"

	"telegram-bot/scraper/belasartes"

	tele "gopkg.in/telebot.v3"
)

func (h *Handler) BelasArtes(c tele.Context) error {
	events, err := belasartes.ScrapeCineBelasArtes()
	if err != nil {
		return c.Send("Sorry, I couldn't fetch the movie list right now.")
	}

	if len(events) == 0 {
		return c.Send("No movies found at the moment.")
	}

	maxItems := 50
	maxItems = min(maxItems, len(events))

	var b strings.Builder
	for i := 0; i < maxItems; i++ {
		ev := events[i]
		b.WriteString(fmt.Sprintf("%d) %s", i+1, ev.Title))
		if ev.Date != "" {
			b.WriteString(fmt.Sprintf(" — %s", ev.Date))
		}
		if i < maxItems-1 {
			b.WriteString("\n\n")
		}
	}

	return c.Send(b.String())
}
