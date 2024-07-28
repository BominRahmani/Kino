package providers

import "github.com/gocolly/colly/v2"

type Movie struct {
	Title     string
	Url       string
	ImageUrl  string
	MediaType string
	Year      string
}

type Provider struct {
	Name      string
	Base_URL  string
	Collector *colly.Collector
}

type Scraper interface {
	Scrape(query string) ([]*Movie, error)
	FormatQuery(query string) string
}
