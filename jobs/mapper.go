package jobs

import (
	"bytes"
	"telegram-bot/message"

	tele "gopkg.in/telebot.v3"
)

// mapMessage converts event content and formatting into Telegram send arguments.
func mapMessage(msg message.Message) (interface{}, *tele.SendOptions) {
	options := &tele.SendOptions{}
	switch msg.Format {
	case message.HTML:
		options.ParseMode = tele.ModeHTML
	case message.Markdown:
		options.ParseMode = tele.ModeMarkdown
	case message.MarkdownV2:
		options.ParseMode = tele.ModeMarkdownV2
	}
	if msg.Document != nil {
		return &tele.Document{File: tele.FromReader(bytes.NewReader(msg.Document.Content)),
			FileName: msg.Document.Name, Caption: msg.Text}, options
	}
	if msg.ImageURL != "" {
		return &tele.Photo{
			File:    tele.FromURL(msg.ImageURL),
			Caption: msg.Text,
		}, options
	}
	return msg.Text, options
}
