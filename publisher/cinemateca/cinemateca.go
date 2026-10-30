package cinemateca

import (
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"telegram-bot/message"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://cinemateca.org.br"

type Event struct {
	id, title, when, url, venue string
	day                         time.Time
}

func (e *Event) ID() string     { return "cinemateca|" + e.id }
func (e *Event) Source() string { return "cinemateca" }

// Digest is one notification containing the entire week's program.
type Digest struct {
	start  time.Time
	events []*Event
}

func (d *Digest) ID() string     { return "cinemateca|week|" + d.start.Format("2006-01-02") }
func (d *Digest) Source() string { return "cinemateca" }
func (d *Digest) Message() message.Message {
	header := fmt.Sprintf("🎬 Cinemateca — %s a %s", d.start.Format("02/01"), d.start.AddDate(0, 0, 6).Format("02/01/2006"))
	var text strings.Builder
	text.WriteString(header)
	if len(d.events) == 0 {
		text.WriteString("\n\nNenhum evento encontrado para esta semana.")
	}
	for _, event := range d.events {
		fmt.Fprintf(&text, "\n\n%s\n📅 %s", event.title, event.when)
		if event.venue != "" {
			fmt.Fprintf(&text, "\n📍 %s", event.venue)
		}
		fmt.Fprintf(&text, "\n%s", event.url)
	}
	content := text.String()
	// Keep long programs in a single Telegram message without losing events.
	if len([]rune(content)) > 4000 {
		return message.Message{Text: header + "\nProgramação completa no arquivo anexo.",
			Document: &message.Document{Name: "cinemateca-" + d.start.Format("2006-01-02") + ".txt", Content: []byte(content)}}
	}
	return message.Message{Text: content, Format: message.PlainText}
}

func Job() (*Digest, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	events, err := getEvents(start)
	if err != nil {
		return nil, err
	}
	return &Digest{start: start, events: events}, nil
}

// getEvents returns events on the next seven calendar days, including start.
func getEvents(start time.Time) ([]*Event, error) {
	loc := start.Location()
	end := start.AddDate(0, 0, 7)
	client := &http.Client{Timeout: 30 * time.Second}
	seen := make(map[string]bool)
	var events []*Event
	month := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, loc)
	for ; month.Before(end); month = month.AddDate(0, 1, 0) {
		doc, err := fetchMonth(client, month)
		if err != nil {
			return nil, err
		}
		for _, event := range parseMonth(doc, start, end) {
			if !seen[event.ID()] {
				seen[event.ID()] = true
				events = append(events, event)
			}
		}
	}
	sort.SliceStable(events, func(i, j int) bool { return events[i].day.Before(events[j].day) })
	return events, nil
}

func fetchMonth(client *http.Client, month time.Time) (*goquery.Document, error) {
	address := baseURL + "/exibicoes/mes/" + month.Format("2006-01") + "/"
	req, err := http.NewRequest(http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (GoScraper)")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("cinemateca calendar: HTTP %d", resp.StatusCode)
	}
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}
	if doc.Find("table.tribe-events-calendar-month").Length() == 0 {
		return nil, fmt.Errorf("cinemateca calendar table missing at %s", address)
	}
	return doc, nil
}

func normalized(text string) string { return strings.Join(strings.Fields(text), " ") }

func parseMonth(doc *goquery.Document, start, end time.Time) []*Event {
	table := doc.Find("table.tribe-events-calendar-month")
	// Multi-day continuations contain only a reference. Find their full entry
	// anywhere in the table, even when the start day is outside our window.
	entries := make(map[string]*goquery.Selection)
	key := func(article *goquery.Selection) string {
		for _, class := range strings.Fields(article.AttrOr("class", "")) {
			if strings.HasPrefix(class, "post-") {
				return class
			}
		}
		return article.Find("a[href]").First().AttrOr("href", "")
	}
	table.Find("article.tribe_events").Each(func(_ int, article *goquery.Selection) {
		id := key(article)
		if entries[id] == nil || article.Find(".tribe-events-calendar-month__calendar-event-tooltip").Length() > 0 {
			entries[id] = article
		}
	})
	seen := make(map[string]bool)
	var events []*Event
	table.Find("td").Each(func(_ int, cell *goquery.Selection) {
		date := cell.Find(".tribe-events-calendar-month__day-date-daynum[datetime]").First().AttrOr("datetime", "")
		day, err := time.ParseInLocation("2006-01-02", date, start.Location())
		if err != nil || day.Before(start) || !day.Before(end) {
			return
		}
		cell.Find("article.tribe_events").Each(func(_ int, ref *goquery.Selection) {
			if !ref.HasClass("cat_sessoes") && !ref.HasClass("cat_post_eventos") &&
				!ref.HasClass("tribe_events_cat-sessoes") && !ref.HasClass("tribe_events_cat-post_eventos") {
				return
			}
			article := entries[key(ref)]
			link := article.Find(".tribe-events-calendar-month__calendar-event-title-link, .tribe-events-calendar-month__multiday-event-hidden-link").First()
			href := link.AttrOr("href", "")
			parsed, err := url.Parse(href)
			if err != nil || href == "" {
				return
			}
			base, _ := url.Parse(baseURL)
			href = base.ResolveReference(parsed).String()
			title := normalized(link.Text())
			if title == "" {
				return
			}
			id := key(ref) + "|" + date
			if ref.HasClass("tribe-events-calendar-month__multiday-event") {
				id = key(ref) // One notification for the entire multi-day event.
			}
			if seen[id] {
				return
			}
			seen[id] = true
			when := normalized(article.Find(".tribe-events-calendar-month__calendar-event-tooltip-datetime time").First().Text())
			if when == "" {
				when = day.Format("02/01/2006")
				clock := normalized(article.Find(".tribe-events-calendar-month__calendar-event-datetime time").First().Text())
				if clock != "" {
					when += " às " + clock
				}
			}
			events = append(events, &Event{id: id, title: title, when: when, url: href,
				venue: normalized(article.Find(".deltec-location-name").First().Text()), day: day})
		})
	})
	return events
}
