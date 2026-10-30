package message

// Format describes how the message text should be parsed.
type Format uint8

const (
	PlainText Format = iota
	HTML
	Markdown
	MarkdownV2
)

// Message is the content of an event, independent of its delivery transport.
type Message struct {
	Text     string
	ImageURL string
	Format   Format
}
