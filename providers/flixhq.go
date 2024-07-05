package providers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

const BASE_FLIXHQ = "https://flixhq.to/"

var (
	collector *colly.Collector
)

func init() {
	collector = colly.NewCollector(
		colly.Async(true),
	)

}

func Scrape(query string) ([]*Movie, error) {

	var kinoList []*Movie
	var mu sync.Mutex
	query = formatQuery(query)
	collector.OnHTML("div.film_list-wrap div div.film-detail", func(h *colly.HTMLElement) {
		kino := &Movie{
			Title:     h.ChildText("h2.film-name"),
			Url:       BASE_FLIXHQ + strings.TrimPrefix(h.ChildAttr("h2.film-name a", "href"), "/"),
			MediaType: h.ChildText("div.fd-infor span.fdi-type"),
			ImageUrl:  "",
			Year:      h.ChildText("div.fd-infor span.fdi-item:first-child"),
		}
		mu.Lock()
		kinoList = append(kinoList, kino)
		mu.Unlock()
	})

	err := collector.Visit(query)
	if err != nil {
		return nil, err
	}
	collector.Wait()
	if len(kinoList) == 0 {
		return nil, fmt.Errorf("No results found for query: %s", query)
	}
	return kinoList, nil
}

func formatQuery(query string) string { //flixhq search replaces all spaces with - and has /search/ format
	return BASE_FLIXHQ + "search/" + strings.ReplaceAll(query, " ", "-")
}
