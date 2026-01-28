
package belasartes

import (
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Event struct {
	Title    string `json:"title"`
	Date     string `json:"date"`
	URL      string `json:"url"`
	Synopsis string `json:"synopsis"`
}

func ScrapeCineBelasArtes() ([]Event, error) {
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

	var events []Event

	doc.Find(".c-movie-card__info-title-content").Each(func(_ int, card *goquery.Selection) {
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

		events = append(events, Event{
			Title:    title,
			Date:     date,
			URL:      href,
			Synopsis: synopsis,
		})
	})

	return events, nil
}
