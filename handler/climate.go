package handler

import (
	"fmt"
	"strings"
	"time"

	"telegram-bot/weather"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

func (h *Handler) SaoPauloClimate(c tele.Context) error {
	h.Logger.Info("command actioned", zap.String("command", "/weather"))
	msg, err := saoPauloClimateMessage(24*time.Hour, 6*time.Hour)
	if err != nil {
		return c.Send(err.Error())
	}
	return c.Send(msg)
}

func (h *Handler) SaoPauloClimateJob(chatID int64) {
	msg, err := saoPauloClimateMessage(24*time.Hour, 6*time.Hour)
	if err != nil {
		h.Bot.Send(tele.ChatID(chatID), err.Error())
		return
	}
	h.Bot.Send(tele.ChatID(chatID), msg)
}

func saoPauloClimateMessage(window time.Duration, block time.Duration) (string, error) {
	data, err := weather.SaoPauloHourlyForecast()
	if err != nil {
		return "", err
	}

	loc, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return "", fmt.Errorf("Sorry, I couldn't load the timezone.")
	}
	now := time.Now().In(loc)
	end := now.Add(window)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("🌤️ Sao Paulo - proximas %.0f horas\n", window.Hours()))

	const layout = "2006-01-02T15:04"
	type bucket struct {
		count       int
		sumTemp     float64
		sumHumidity float64
		sumPrecip   float64
		blockStart  time.Time
		blockEnd    time.Time
	}

	blockCount := int(window / block)
	buckets := make([]bucket, blockCount)
	for i := 0; i < blockCount; i++ {
		start := now.Add(time.Duration(i) * block)
		buckets[i] = bucket{
			blockStart: start,
			blockEnd:   start.Add(block),
		}
	}

	for i, t := range data.Hourly.Time {
		if i >= len(data.Hourly.Temperature2m) || i >= len(data.Hourly.Humidity2m) || i >= len(data.Hourly.Precipitation) {
			break
		}
		ts, err := time.ParseInLocation(layout, t, loc)
		if err != nil {
			continue
		}
		if ts.Before(now) || !ts.Before(end) {
			continue
		}
		index := int(ts.Sub(now) / block)
		if index < 0 || index >= blockCount {
			continue
		}
		buckets[index].count++
		buckets[index].sumTemp += data.Hourly.Temperature2m[i]
		buckets[index].sumHumidity += data.Hourly.Humidity2m[i]
		buckets[index].sumPrecip += data.Hourly.Precipitation[i]
	}

	line := 0
	for _, bucket := range buckets {
		if bucket.count == 0 {
			continue
		}
		line++
		avgTemp := bucket.sumTemp / float64(bucket.count)
		avgHumidity := bucket.sumHumidity / float64(bucket.count)
		b.WriteString(fmt.Sprintf(
			"%d. %s - %.1f%s, %.0f%s, %.1f%s\n",
			line,
			bucket.blockStart.Format("Mon 15h"),
			avgTemp,
			data.HourlyUnits.Temperature2m,
			avgHumidity,
			data.HourlyUnits.Humidity2m,
			bucket.sumPrecip,
			data.HourlyUnits.Precipitation,
		))
	}

	if line == 0 {
		return "", fmt.Errorf("Sorry, I couldn't find upcoming temperatures.")
	}

	return strings.TrimRight(b.String(), "\n"), nil
}
