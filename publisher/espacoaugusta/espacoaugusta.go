package espacoaugusta

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Event struct {
	title    string
	director string
	info     string
	poster   string
}

func (e *Event) ID() string {
	return fmt.Sprintf("espacoaugusta-%s", e.title)
}

func (e *Event) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("<b>%s</b>\n", e.title))
	b.WriteString(fmt.Sprintf("<i>%s</i>\n", e.director))
	b.WriteString(e.info)
	return b.String()
}

func (e *Event) Source() string {
	return "espacoaugusta"
}

func (e *Event) Image() string {
	return e.poster
}

func GetEvents() ([]*Event, error) {
	url := "https://espacopetrobrasdecinema.com.br/"
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (GoScraper)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	var events []*Event
	processed := make(map[string]struct{})

	doc.Find(".fusion-post-cards .fusion-grid-column").Each(func(_ int, card *goquery.Selection) {
		title := strings.TrimSpace(card.Find(".fusion-title-heading").First().Text())
		if title == "" || title == "Programação" {
			return
		}

		if _, ok := processed[title]; ok {
			return
		}
		processed[title] = struct{}{}

		poster, _ := card.Find("img").First().Attr("data-orig-src")

		headings := card.Find("h4")
		var director, info string
		headings.Each(func(_ int, h *goquery.Selection) {
			text := strings.TrimSpace(h.Text())
			if strings.HasPrefix(text, "Dirigido por") {
				director = text
			} else if director != "" && info == "" && !strings.HasPrefix(text, "Não recomendado") && text != "" {
				info = text
			}
		})

		events = append(events, &Event{
			title:    title,
			director: director,
			info:     info,
			poster:   poster,
		})
	})

	return events, nil
}
