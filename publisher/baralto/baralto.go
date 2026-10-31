package baralto

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"telegram-bot/message"
)

const (
	apiURL  = "https://api.meaple.com.br/v1/channels/"
	pageURL = "https://meaple.com.br/baralto"
)

type event struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Slug       string     `json:"slug"`
	StartsAt   time.Time  `json:"startsAt"`
	Timezone   string     `json:"timezone"`
	CanceledAt *time.Time `json:"canceledAt"`
	Status     string     `json:"status"`
}

// Digest contains the next seven days of Bar Alto events in one notification.
type Digest struct {
	start  time.Time
	events []event
}

func (d *Digest) ID() string     { return "baralto|week|" + d.start.Format("2006-01-02") }
func (d *Digest) Source() string { return "baralto" }
func (d *Digest) Message() message.Message {
	header := fmt.Sprintf("🎵 Bar Alto — %s a %s", d.start.Format("02/01"), d.start.AddDate(0, 0, 6).Format("02/01/2006"))
	var text strings.Builder
	text.WriteString(header)
	if len(d.events) == 0 {
		text.WriteString("\n\nNenhum evento encontrado para esta semana.")
	}
	for _, event := range d.events {
		fmt.Fprintf(&text, "\n\n%s\n📅 %s\n%s/%s", event.Name,
			event.StartsAt.Format("02/01 às 15:04 MST"), pageURL, url.PathEscape(event.Slug))
	}
	content := text.String()
	if len([]rune(content)) > 4000 {
		return message.Message{Text: header + "\nProgramação completa no arquivo anexo.",
			Document: &message.Document{Name: "baralto-" + d.start.Format("2006-01-02") + ".txt", Content: []byte(content)}}
	}
	return message.Message{Text: content, Format: message.PlainText}
}

func Job() (*Digest, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	client := &http.Client{Timeout: 30 * time.Second}
	var channel struct {
		Channel struct {
			ID string `json:"id"`
		} `json:"channel"`
	}
	if err := getJSON(client, apiURL+"baralto", &channel); err != nil {
		return nil, err
	}
	if channel.Channel.ID == "" {
		return nil, fmt.Errorf("baralto: missing channel ID")
	}
	var response struct {
		Events []event `json:"events"`
	}
	if err := getJSON(client, apiURL+url.PathEscape(channel.Channel.ID)+"/events?type=FUTURE", &response); err != nil {
		return nil, err
	}
	if response.Events == nil {
		return nil, fmt.Errorf("baralto: missing events in response")
	}
	end := start.AddDate(0, 0, 7).Format("2006-01-02")
	seen := make(map[string]bool)
	digest := &Digest{start: start}
	for _, event := range response.Events {
		if event.CanceledAt != nil || event.Status != "PUBLISHED" {
			continue
		}
		if event.ID == "" || event.Name == "" || event.Slug == "" || event.StartsAt.IsZero() {
			return nil, fmt.Errorf("baralto: incomplete event in response")
		}
		loc := time.Local
		if event.Timezone != "" {
			var err error
			loc, err = time.LoadLocation(event.Timezone)
			if err != nil {
				return nil, fmt.Errorf("baralto event timezone: %w", err)
			}
		}
		event.StartsAt = event.StartsAt.In(loc)
		// Compare venue calendar dates so late-night events are not shifted
		// into a different day when the bot's system timezone is UTC.
		day := event.StartsAt.Format("2006-01-02")
		if day < start.Format("2006-01-02") || day >= end || seen[event.ID] {
			continue
		}
		seen[event.ID] = true
		digest.events = append(digest.events, event)
	}
	sort.SliceStable(digest.events, func(i, j int) bool {
		return digest.events[i].StartsAt.Before(digest.events[j].StartsAt)
	})
	return digest, nil
}

func getJSON(client *http.Client, address string, result interface{}) error {
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (GoScraper)")
	req.Header.Set("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("baralto: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("baralto: HTTP %d from %s", resp.StatusCode, address)
	}
	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("baralto response: %w", err)
	}
	return nil
}
