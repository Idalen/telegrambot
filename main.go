package main

import (
	"log"
	"os"
	"time"

	"telegram-bot/handler"
	"telegram-bot/register"

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

	bot.Handle("/belasartes", h.BelasArtes)
	bot.Handle("/startbelasartes", h.StartBelasArtes)
	bot.Handle("/stopbelasartes", h.StopBelasArtes)

	log.Println("Bot is running...")
	bot.Start()
}
