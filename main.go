package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

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

	// /start command
	bot.Handle("/start", func(c tele.Context) error {
		return c.Send("Hello! 👋 I am your Go Telegram bot.")
	})

	bot.Handle("/days", func(c tele.Context) error {
		startDate := time.Date(2023, 8, 6, 0, 0, 0, 0, time.UTC)
		now := time.Now().UTC()

		days := int(now.Sub(startDate).Hours() / 24)

		return c.Send(fmt.Sprintf(
			"Hello! 👋\nIt has been %d days since 2023-08-06.",
			days,
		))
	})

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
