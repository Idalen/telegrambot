package belasartes

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Event struct {
	title    string
	date     string
	url      string
	synopsis string
}

func (e *Event) ID() string {
	return e.Source()+e.title+e.date
}

func (e *Event) String() string {
	return fmt.Sprintf(
		"🎬 <b>%s</b>\n"+
			"📅 <i>Quando: %s</i>\n\n"+
			"📝 %s\n\n"+
			"🔗 <a href=\"%s\">Ver detalhes</a>",
		e.title,
		e.date,
		e.synopsis,
		e.url,
	)
}

func (e *Event) Source() string {
	return "belasartes"
}

func (e *Event) Image() string {
	return ""
}

func GetEvents() ([]*Event, error) {
	url := "https://www.cinebelasartes.com.br/programacao-especial/"
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

	doc.Find(".c-movie-card").Each(func(_ int, card *goquery.Selection) {
		title := strings.TrimSpace(
			card.Find(".c-movie-card__info-title").Text(),
		)

		date := strings.TrimSpace(
			card.Find(".c-movie-card__info-subtitle").Text(),
		)

		synopsis := strings.TrimSpace(
			card.Find(".c-movie-card__info-synopsis").Text(),
		)

		href, _ := card.Find("a").First().Attr("href")

		if title == "" {
			return
		}

		if strings.HasPrefix(href, "/") {
			href = "https://www.cinebelasartes.com.br" + href
		}

		events = append(events, &Event{
			title:    title,
			date:     date,
			url:      href,
			synopsis: synopsis,
		})
	})

	return events, nil
}
