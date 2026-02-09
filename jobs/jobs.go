package jobs

import (
	"context"
	"fmt"
	"time"

	tele "gopkg.in/telebot.v3"
	"telegram-bot/register"
)

type Orchestrator struct {
	Register *register.JobRegistry
}

func (o *Orchestrator) Start(c tele.Context, method string, job func(int64), frequency time.Duration) error {
	chatID := c.Chat().ID

	started := o.Register.StartOnce(chatID, method, func(ctx context.Context) {
		ticker := time.NewTicker(frequency)
		defer ticker.Stop()

		job(chatID)

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				job(chatID)
			}
		}
	})

	if !started {
		return c.Send(fmt.Sprintf("%s is already running", method))
	}

	return c.Send(fmt.Sprintf("Started %s", method))
}

func (o *Orchestrator) Stop(c tele.Context, method string) error {
	if !o.Register.Stop(c.Chat().ID, method) {
		return fmt.Errorf("%s method could not be stopped", method)
	}

	return nil
}
