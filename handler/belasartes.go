package handler

import (
	"fmt"
	"html"
	"strings"

	"telegram-bot/scraper/belasartes"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) BelasArtes(c tele.Context) error {
	h.Logger.Info("command actioned", zap.String("command", "/belasartes"))
	events, err := belasartes.GetEvents()
	if err != nil {
		return c.Send("Sorry, I couldn't fetch the movie list right now.")
	}

	if len(events) == 0 {
		return c.Send("No movies found at the moment.")
	}

	maxItems := 50
	maxItems = min(maxItems, len(events))

	var b strings.Builder
	b.WriteString("🎬 <b>Programacao Belas Artes</b>\n")
	for i := 0; i < maxItems; i++ {
		ev := events[i]
		title := html.EscapeString(ev.Title)
		date := html.EscapeString(ev.Date)
		url := html.EscapeString(ev.URL)
		b.WriteString(fmt.Sprintf("\n%d) <b>%s</b>", i+1, title))
		if ev.Date != "" {
			b.WriteString(fmt.Sprintf("\n   <i>%s</i>", date))
		}
		if ev.URL != "" {
			b.WriteString(fmt.Sprintf("\n   <a href=\"%s\">Ver detalhes</a>", url))
		}
		if i < maxItems-1 {
			b.WriteString("\n")
		}
	}

	return c.Send(b.String(), tele.ModeHTML)
}

func (h *Handler) BelasArtesJob(chatID int64) {
	events, _ := belasartes.GetEvents()
	if len(events) == 0 {
		return
	}

	toSend := h.Store.MarkSeenBelasArtes(chatID, events)

	for _, event := range toSend {
		h.Bot.Send(
			tele.ChatID(chatID),
			formatEventMessage(event),
			tele.ModeMarkdown,
		)
	}
}

func formatEventMessage(e belasartes.Event) string {
	return fmt.Sprintf(
		"🎬 *%s*\n"+
			"📅 *Quando:* %s\n\n"+
			"📝 %s\n\n"+
			"🔗 %s",
		e.Title,
		e.Date,
		e.Synopsis,
		e.URL,
	)
}
