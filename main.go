package main

import (
	"os"
	"time"

	"telegram-bot/handler"
	"telegram-bot/jobs"
	"telegram-bot/router"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		logger.Fatal("TELEGRAM_BOT_TOKEN is not set")
	}

	pref := tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	bot, err := tele.NewBot(pref)
	if err != nil {
		logger.Fatal("failed to create bot", zap.Error(err))
	}

	subscriptor := jobs.New(bot, logger)

	h := &handler.Handler{
		Bot:    bot,
		Logger: logger,
	}
	r := &router.Router{
		Subscriptor: subscriptor,
		Logger:      logger,
	}

	bot.Handle("/ping", h.Start)
	bot.Handle("/jobs", r.HandleJobs)

	bot.Handle(tele.OnText, r.HandleText)

	logger.Info("bot is running")
	bot.Start()
}
