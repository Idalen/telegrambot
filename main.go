package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"telegram-bot/handler"
	"telegram-bot/register"
	"telegram-bot/scraper/belasartes"

	tele "gopkg.in/telebot.v3"
)

func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}

	h := &handler.Handler{
		Bot: bot,
		Register: register.New(),
	}

	bot.Handle("/start", h.Start)

	bot.Handle("/dayssincewemet", h.DaysSinceWeMet)
	bot.Handle("/stopdayssincewemet", h.StopDaysSinceWeMet)

	bot.Handle("/movies", func(c tele.Context) error {
		events, err := belasartes.ScrapeCineBelasArtes()
		if err != nil {
			log.Printf("scrape cinebelasartes: %v", err)
			return c.Send("Sorry, I couldn't fetch the movie list right now.")
		}

		if len(events) == 0 {
			return c.Send("No movies found at the moment.")
		}

		maxItems := 10
		if len(events) < maxItems {
			maxItems = len(events)
		}

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
	})

	log.Println("Bot is running...")
	bot.Start()
}
