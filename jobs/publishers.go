package jobs

import (
	"sync"
	"time"

	"telegram-bot/constants"
	"telegram-bot/publisher/belasartes"
	"telegram-bot/publisher/climate"
	"telegram-bot/publisher/daystogether"
	"telegram-bot/publisher/espacoaugusta"

	"go.uber.org/zap"
)

// Gather polls one source and publishes its events until done is closed.
func Gather(events chan<- Event, done <-chan struct{}, job func() ([]Event, error), frequency time.Duration, logger *zap.Logger) {
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			batch, err := job()
			if err != nil {
				logger.Error("job failed", zap.Error(err))
				continue
			}
			for _, event := range batch {
				select {
				case <-done:
					return
				case events <- event:
				}
			}
		}
	}
}

// RunPublishers starts all information gatherers at application startup.
func RunPublishers(events chan<- Event, done <-chan struct{}, logger *zap.Logger) {
	var wg sync.WaitGroup
	start := func(name string, job func() ([]Event, error), frequency time.Duration) {
		wg.Add(1)
		go func() {
			defer wg.Done()
			Gather(events, done, job, frequency, logger.With(zap.String("source", name)))
		}()
	}
	start(constants.MethodBelasArtes, func() ([]Event, error) {
		events, err := belasartes.GetEvents()
		if err != nil {
			return nil, err
		}
		result := make([]Event, len(events))
		for i, e := range events {
			result[i] = e
		}
		return result, nil
	}, 1*time.Hour)

	start(constants.MethodWeather, func() ([]Event, error) {
		event, err := climate.Job()
		if err != nil {
			return nil, err
		}
		return []Event{event}, nil
	}, 24*time.Hour)

	start(constants.MethodDaysTogether, func() ([]Event, error) {
		event, err := daystogether.Job()
		if err != nil {
			return nil, err
		}
		return []Event{event}, nil
	}, 24*time.Hour)

	start(constants.MethodEspacoAugusta, func() ([]Event, error) {
		events, err := espacoaugusta.GetEvents()
		if err != nil {
			return nil, err
		}
		result := make([]Event, len(events))
		for i, e := range events {
			result[i] = e
		}
		return result, nil
	}, 1*time.Hour)

	wg.Wait()
}
