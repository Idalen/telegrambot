package main

import (
	"fmt"
	"log"
	"os"
	"time"

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

	// Echo any text message
	bot.Handle(tele.OnText, func(c tele.Context) error {
		return c.Send("You said: " + c.Text())
	})

	log.Println("Bot is running...")
	bot.Start()
}
