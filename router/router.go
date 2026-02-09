package router

import (
	"fmt"
	"strings"
	"time"

	"telegram-bot/constants"
	"telegram-bot/handler"
	"telegram-bot/jobs"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

type Router struct {
	Orchestrator *jobs.Orchestrator
	Handler      *handler.Handler
	Logger       *zap.Logger
}

func (r *Router) HandleText(c tele.Context) error {
	if err := r.handleStartText(c); err != nil {
		return err
	}
	return r.handleStopText(c)
}

func (r *Router) HandleJobs(c tele.Context) error {
	r.Logger.Info("command actioned", zap.String("command", "/jobs"))
	methods := r.Orchestrator.Register.Methods(c.Chat().ID)
	if len(methods) == 0 {
		return c.Send("No active jobs.")
	}

	var b strings.Builder
	b.WriteString("Active jobs:\n")
	for i, method := range methods {
		b.WriteString(fmt.Sprintf("%d. %s\n", i+1, method))
	}
	return c.Send(strings.TrimRight(b.String(), "\n"))
}

func (r *Router) handleStartText(c tele.Context) error {
	text := strings.TrimSpace(c.Text())
	if !strings.HasPrefix(text, constants.CommandStart) {
		return nil
	}

	parts := strings.Fields(text)
	if len(parts) < 2 {
		return nil
	}

	switch parts[1] {
	case constants.CommandBelasArtes:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStart),
			zap.String("target", constants.CommandBelasArtes),
		)
		return r.Orchestrator.Start(c, constants.MethodBelasArtes, r.Handler.BelasArtesJob, 1*time.Hour)
	case constants.CommandWeather:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStart),
			zap.String("target", constants.CommandWeather),
		)
		return r.Orchestrator.Start(c, constants.MethodWeather, r.Handler.SaoPauloClimateJob, 24*time.Hour)
	case constants.CommandDaysTogether:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStart),
			zap.String("target", constants.CommandDaysTogether),
		)
		return r.Orchestrator.Start(c, constants.MethodDaysTogether, r.Handler.DaysTogetherJob, 24*time.Hour)
	}

	return nil
}

func (r *Router) handleStopText(c tele.Context) error {
	text := strings.TrimSpace(c.Text())
	if !strings.HasPrefix(text, constants.CommandStop) {
		return nil
	}

	parts := strings.Fields(text)
	if len(parts) < 2 {
		return nil
	}

	switch parts[1] {
	case constants.CommandBelasArtes:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStop),
			zap.String("target", constants.CommandBelasArtes),
		)
		if err := r.Orchestrator.Stop(c, constants.MethodBelasArtes); err != nil {
			return c.Send(err.Error())
		}
		return c.Send("Stopped belasartes updates.")
	case constants.CommandWeather:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStop),
			zap.String("target", constants.CommandWeather),
		)
		if err := r.Orchestrator.Stop(c, constants.MethodWeather); err != nil {
			return c.Send(err.Error())
		}
		return c.Send("Stopped weather updates.")
	case constants.CommandDaysTogether:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStop),
			zap.String("target", constants.CommandDaysTogether),
		)
		if err := r.Orchestrator.Stop(c, constants.MethodDaysTogether); err != nil {
			return c.Send(err.Error())
		}
		return c.Send("Stopped daystogether updates.")
	}

	return nil
}
