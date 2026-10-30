package daystogether

import (
	"fmt"
	"telegram-bot/message"
	"time"

	"telegram-bot/constants"
)

const _message = "Hoje os momorzinhos fazem %d dias juntos! 🥰❤️"

type Event struct {
	message string
}

func (e *Event) ID() string {
	return "daystogether-" + time.Now().Format("2006-01-02")
}

func (e *Event) Message() message.Message {
	return message.Message{Format: message.PlainText, Text: e.message}
}

func (e *Event) Source() string {
	return constants.MethodDaysTogether
}

func Job() (*Event, error) {
	days := daysTogether()
	return &Event{message: fmt.Sprintf(_message, days)}, nil
}

func daysTogether() int {
	dateWeMet := time.Date(2023, 8, 6, 0, 0, 0, 0, time.Local)
	now := time.Now()
	return int(now.Sub(dateWeMet).Hours() / 24)
}
