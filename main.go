package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"telegram-bot/constants"
	"telegram-bot/jobs"

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

	events := make(chan jobs.Event)
	done := make(chan struct{})
	broadcaster := jobs.New(bot, logger)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() { defer wg.Done(); jobs.RunPublishers(events, done, logger) }()
	go func() { defer wg.Done(); broadcaster.Run(events, done) }()
	defer wg.Wait()
	defer close(done)

	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	go func() {
		select {
		case <-signals:
			bot.Stop()
		case <-done:
		}
	}()

	bot.Handle(constants.CommandSubscribe, func(c tele.Context) error {
		broadcaster.Subscribe(c.Chat().ID)
		return c.Send("Subscribed! You'll receive updates from all topics.")
	})

	logger.Info("bot is running")
	bot.Start()
}
