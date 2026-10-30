package router

import (
	"reflect"
	"sort"
	"testing"

	"telegram-bot/constants"
	"telegram-bot/jobs"

	"go.uber.org/zap"
	tele "gopkg.in/telebot.v3"
)

type subscriptionContext struct {
	tele.Context
	chat     tele.Chat
	messages []interface{}
}

func (c *subscriptionContext) Chat() *tele.Chat { return &c.chat }
func (c *subscriptionContext) Send(what interface{}, opts ...interface{}) error {
	c.messages = append(c.messages, what)
	return nil
}

func TestSubscribeStartsAllTopicsAndCanBeRepeated(t *testing.T) {
	logger := zap.NewNop()
	s := jobs.New(nil, logger)
	defer s.Close()
	r := &Router{Subscriptor: s, Logger: logger}
	want := []string{constants.MethodBelasArtes, constants.MethodWeather, constants.MethodDaysTogether, constants.MethodEspacoAugusta}
	sort.Strings(want)

	for _, chatID := range []int64{1, 1, 2} {
		c := &subscriptionContext{chat: tele.Chat{ID: chatID}}
		if err := r.HandleSubscribe(c); err != nil {
			t.Fatal(err)
		}
		if len(c.messages) != 1 || c.messages[0] != "Subscribed! You'll receive updates from all topics." {
			t.Fatalf("unexpected confirmation: %v", c.messages)
		}
		got := s.Register.Methods()
		sort.Strings(got)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("active topics = %v, want %v", got, want)
		}
	}
}
