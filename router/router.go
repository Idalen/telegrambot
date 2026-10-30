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

// HandleSubscribe subscribes the chat to every supported topic.
func (r *Router) HandleSubscribe(c tele.Context) error {
	var errs []string
	errs = append(errs, r.startBelasArtes(c)...)
	errs = append(errs, r.startWeather(c)...)
	errs = append(errs, r.startDaysTogether(c)...)
	errs = append(errs, r.startEspacoAugusta(c)...)
	if len(errs) > 0 {
		return c.Send("Could not subscribe to all topics:\n" + strings.Join(errs, "\n"))
	}
	return c.Send("Subscribed! You'll receive updates from all topics.")
}

func (r *Router) startBelasArtes(c tele.Context) []string {
	r.Logger.Info(
		"command actioned",
		zap.String("command", constants.CommandSubscribe),
		zap.String("target", constants.MethodBelasArtes),
	)
	if err := r.Subscriptor.Start(c, constants.MethodBelasArtes, func() ([]jobs.Event, error) {
		events, err := belasartes.GetEvents()
		if err != nil {
			return nil, err
		}
		result := make([]jobs.Event, len(events))
		for i, e := range events {
			result[i] = e
		}
		return result, nil
	}, 1*time.Hour); err != nil {
		return []string{err.Error()}
	}
	return nil
}

func (r *Router) startWeather(c tele.Context) []string {
	r.Logger.Info(
		"command actioned",
		zap.String("command", constants.CommandSubscribe),
		zap.String("target", constants.MethodWeather),
	)
	if err := r.Subscriptor.Start(c, constants.MethodWeather, func() ([]jobs.Event, error) {
		event, err := climate.Job()
		if err != nil {
			return nil, err
		}
		return []jobs.Event{event}, nil
	}, 24*time.Hour); err != nil {
		return []string{err.Error()}
	}
	return nil
}

func (r *Router) startDaysTogether(c tele.Context) []string {
	r.Logger.Info(
		"command actioned",
		zap.String("command", constants.CommandSubscribe),
		zap.String("target", constants.MethodDaysTogether),
	)
	if err := r.Subscriptor.Start(c, constants.MethodDaysTogether, func() ([]jobs.Event, error) {
		event, err := daystogether.Job()
		if err != nil {
			return nil, err
		}
		return []jobs.Event{event}, nil
	}, 24*time.Hour); err != nil {
		return []string{err.Error()}
	}
	return nil
}

func (r *Router) startEspacoAugusta(c tele.Context) []string {
	r.Logger.Info(
		"command actioned",
		zap.String("command", constants.CommandSubscribe),
		zap.String("target", constants.MethodEspacoAugusta),
	)
	if err := r.Subscriptor.Start(c, constants.MethodEspacoAugusta, func() ([]jobs.Event, error) {
		events, err := espacoaugusta.GetEvents()
		if err != nil {
			return nil, err
		}
		result := make([]jobs.Event, len(events))
		for i, e := range events {
			result[i] = e
		}
		return result, nil
	}, 1*time.Hour); err != nil {
		return []string{err.Error()}
	}
	return nil
}
