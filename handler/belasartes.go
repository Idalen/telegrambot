package handler

import (
	"context"
	"fmt"
	"html"
	"strings"
	"time"

	"telegram-bot/scraper/belasartes"

	tele "gopkg.in/telebot.v3"
)

const _belasArtesMethod = "belasartes"

func (h *Handler) BelasArtes(c tele.Context) error {
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

func (h *Handler) StartBelasArtes(c tele.Context) error {
	chatID := c.Chat().ID

	started := h.Register.StartOnce(chatID, _belasArtesMethod, func(ctx context.Context) {
		h.belasArtesJob(ctx, c.Chat().ID)
	})

	if !started {
		return c.Send(fmt.Sprintf("Already running: %s", _belasArtesMethod))
	}
	return c.Send(fmt.Sprintf("Started %s", _belasArtesMethod))
}

func (h Handler) StopBelasArtes(c tele.Context) error {
	if !h.Register.Stop(c.Chat().ID, _belasArtesMethod) {
		return fmt.Errorf("%s method could not be stopped", _belasArtesMethod)
	}

	return nil
}

func (h *Handler) belasArtesJob(ctx context.Context, chatID int64) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	seenEvents := map[belasartes.Event]struct{}{}

	sendNewEvents := func() {
		events, _ := belasartes.GetEvents()
		for _, event := range events {
			if _, ok := seenEvents[event]; !ok {
				h.Bot.Send(
					tele.ChatID(chatID),
					formatEventMessage(event),
					tele.ModeMarkdown,
				)
				seenEvents[event] = struct{}{}
			}
		}
	}

	sendNewEvents()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sendNewEvents()
		}
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
