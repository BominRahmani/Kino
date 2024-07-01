package providers

import (
	"github.com/gocolly/colly/v2"
)

const BASE_FLIXHQ = "https://flixhq.to/"

func Scrape() {
	// Parse the BASE_FLIXHQ URL
	c := colly.NewCollector()
	var kinoList []*Movie
	c.OnHTML("div.film_list-wrap div div.film-detail", func(h *colly.HTMLElement) {
		kino := &Movie{}
		kino.title = h.ChildText("h2.film-name")
		kino.url = h.ChildAttr("h2.film-name a", "href")
		kino.mediaType = h.ChildText("div.fd-infor span.fdi-type")
		year := h.ChildText("div.fd-infor span.fdi-item:first-child")
		kino.title = kino.title + " " + year
		kinoList = append(kinoList, kino)
	})
	c.Visit("https://flixhq.to/search/life-of-pi")
}
