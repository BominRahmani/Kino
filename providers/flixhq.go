package providers

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gocolly/colly/v2"
)

const BASE_FLIXHQ = "https://flixhq.to/"

const MAX_CONCURRENT_REQUESTS = 5

func formatQuery(query string) string { //flixhq search replaces all spaces with - and has /search/ format
	return BASE_FLIXHQ + "search/" + strings.ReplaceAll(query, " ", "-")
}

func Scrape(query string) ([]*Movie, error) {
	// Parse the BASE_FLIXHQ URL
	c := colly.NewCollector(
		colly.Async(true),
		colly.MaxDepth(2),
	)
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: MAX_CONCURRENT_REQUESTS,
	})

	var kinoList []*Movie
	var mu sync.Mutex
	query = formatQuery(query)
	c.OnHTML("div.film_list-wrap div div.film-detail", func(h *colly.HTMLElement) {
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

	err := c.Visit(fmt.Sprintf(query))
	if err != nil {
		return nil, err
	}
	c.Wait()
	if len(kinoList) == 0 {
		return nil, fmt.Errorf("No results found for query: %s", query)
	}
	return kinoList, nil
}
