package router

import (
	"strings"
	"time"

	"telegram-bot/constants"
	"telegram-bot/jobs"
	"telegram-bot/publisher/belasartes"
	"telegram-bot/publisher/climate"
	"telegram-bot/publisher/daystogether"
	"telegram-bot/publisher/espacoaugusta"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

type Router struct {
	Subscriptor *jobs.Subscriptor
	Logger      *zap.Logger
}

func (r *Router) HandleText(c tele.Context) error {
	if err := r.handleStartText(c); err != nil {
		return err
	}
	return r.handleStopText(c)
}

func (r *Router) HandleJobs(c tele.Context) error {
	r.Logger.Info("command actioned", zap.String("command", "/jobs"))
	methods := r.Subscriptor.Register.Methods()
	if len(methods) == 0 {
		return c.Send("No active jobs.")
	}

	var b strings.Builder
	b.WriteString("Active jobs:\n")
	for i, method := range methods {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- " + method)
	}
	return c.Send(b.String())
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
		return r.Subscriptor.Start(c, constants.MethodBelasArtes, func() ([]jobs.Event, error) {
			events, err := belasartes.GetEvents()
			if err != nil {
				return nil, err
			}
			result := make([]jobs.Event, len(events))
			for i, e := range events {
				result[i] = e
			}
			return result, nil
		}, 1*time.Hour)
	case constants.CommandWeather:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStart),
			zap.String("target", constants.CommandWeather),
		)
		return r.Subscriptor.Start(c, constants.MethodWeather, func() ([]jobs.Event, error) {
			event, err := climate.Job()
			if err != nil {
				return nil, err
			}
			return []jobs.Event{event}, nil
		}, 24*time.Hour)
	case constants.CommandDaysTogether:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStart),
			zap.String("target", constants.CommandDaysTogether),
		)
		return r.Subscriptor.Start(c, constants.MethodDaysTogether, func() ([]jobs.Event, error) {
			event, err := daystogether.Job()
			if err != nil {
				return nil, err
			}
			return []jobs.Event{event}, nil
		}, 24*time.Hour)
	case constants.CommandEspacoAugusta:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStart),
			zap.String("target", constants.CommandEspacoAugusta),
		)
		return r.Subscriptor.Start(c, constants.MethodEspacoAugusta, func() ([]jobs.Event, error) {
			events, err := espacoaugusta.GetEvents()
			if err != nil {
				return nil, err
			}
			result := make([]jobs.Event, len(events))
			for i, e := range events {
				result[i] = e
			}
			return result, nil
		}, 1*time.Hour)
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
		if err := r.Subscriptor.Stop(c, constants.MethodBelasArtes); err != nil {
			return c.Send(err.Error())
		}
		return c.Send("Stopped belasartes updates.")
	case constants.CommandWeather:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStop),
			zap.String("target", constants.CommandWeather),
		)
		if err := r.Subscriptor.Stop(c, constants.MethodWeather); err != nil {
			return c.Send(err.Error())
		}
		return c.Send("Stopped weather updates.")
	case constants.CommandDaysTogether:
		r.Logger.Info(
			"command actioned",
			zap.String("command", constants.CommandStop),
			zap.String("target", constants.CommandDaysTogether),
		)
		if err := r.Subscriptor.Stop(c, constants.MethodDaysTogether); err != nil {
			return c.Send(err.Error())
		}
		return c.Send("Stopped daystogether updates.")
	}

	return nil
}
