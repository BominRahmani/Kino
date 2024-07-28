package providers

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"

	"github.com/gocolly/colly/v2"
)

type FlixHQProvider struct {
	Provider
}

func NewFlixHQProvider() *FlixHQProvider {
	collector := colly.NewCollector()
	collector.Limit(&colly.LimitRule{
		DomainGlob: "*",
	})

	return &FlixHQProvider{
		Provider: Provider{
			Name:      "FlixHQ",
			Base_URL:  "https://flixhq.to/",
			Collector: collector,
		},
	}
}

func (f *FlixHQProvider) Scrape(query string) ([]*Movie, error) {
	var kinoList []*Movie
	query = f.formatQuery(query)

	f.Collector.OnHTML(".film_list-wrap", func(h *colly.HTMLElement) {
		h.ForEach(".flw-item", func(_ int, el *colly.HTMLElement) {
			imgSrc := el.ChildAttr(".film-poster-img", "src")
			filmYear := el.Attr(".film-detail")
			if imgSrc == "" {
				imgSrc = el.ChildAttr(".film-poster-img", "data-src")
			}
			title := el.ChildAttr(".film-poster-img", "title")
			href := el.ChildAttr(".film-poster-ahref", "href")
			kino := &Movie{
				Title:     title,
				Url:       strings.TrimRight(f.Base_URL, "/") + href,
				MediaType: "Movie",
				ImageUrl:  imgSrc,
				Year:      filmYear,
			}
			kinoList = append(kinoList, kino)
		})
	})

	err := f.Collector.Visit(query)
	if err != nil {
		return nil, err
	}

	if len(kinoList) == 0 {
		return nil, fmt.Errorf("No results found for query: %s", query)
	}
	f.Collector.Wait()

	return kinoList, nil
}

func (f *FlixHQProvider) formatQuery(query string) string {
	return f.Base_URL + "search/" + strings.ReplaceAll(query, " ", "-")
}

// GetEmbedLink returns the embeded link for the video streams
// in the form of /watch-movie/<watch-<name-of-movie-split-by-hyphen-delim>-<imdb-identifier>.<unique-id>
func (f *FlixHQProvider) GetEmbedLink(url string) string {
	parts := strings.Split(url, "-")
	movieId := parts[len(parts)-1]
	f.Collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("Content-Type", "application/json")
	})
	var embededLink string
	ajaxUrl := f.Base_URL + "ajax/episode/list/" + movieId
	f.Collector.OnHTML(".server-select .nav-item:first-child a[href]", func(h *colly.HTMLElement) {
		embededLink = h.Attr("href")
	})

	f.Collector.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "Failed")
	})

	err := f.Collector.Visit(ajaxUrl)
	if err != nil {
		fmt.Println("Error visiting URL:")
	}
	f.Collector.Wait()

	return embededLink
}

func (f *FlixHQProvider) GetRabbitID(url string) string {
	parts := strings.Split(url, ".")
	uniqueId := parts[len(parts)-1]
	requestUrl := f.Base_URL + path.Join("ajax", "episode", "sources", uniqueId)
	f.Collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("Content-Type", "application/json")
	})

	var rabbitID string

	// grab the rabbit id
	f.Collector.OnResponse(func(r *colly.Response) {
		var result map[string]interface{}

		err := json.Unmarshal(r.Body, &result)
		if err != nil {
			fmt.Printf("Unable to Unmarshal JSON when parsing for rabbitID")
		}

		// Check if 'link' exists and is a string
		if link, ok := result["link"].(string); ok {
			if strings.Contains(link, "rabbitstream.net") {
				parts := strings.Split(link, "/")
				rabbitID = parts[len(parts)-1]
			}
		}
	})

	err := f.Collector.Visit(requestUrl)
	if err != nil {
		fmt.Println("Error: was unsuccesful in making get request to grab rabbitID", err)
	}

	rabbitID = strings.TrimRight(rabbitID, "?z=")
	return rabbitID
}

// extractRabbitWasm makes use the rabbit wasm extractor to decrypt
func (f *FlixHQProvider) extractRabbitWasm(id string) string {
	var movieLink string
	decryptUrl := "https://lobster-decryption.netlify.app/decrypt?id=" + id
	f.Collector.OnRequest(func(r *colly.Request) {
		r.Headers.Set("X-Requested-With", "XMLHttpRequest")
		r.Headers.Set("Content-Type", "application/json")
	})

	f.Collector.OnResponse(func(r *colly.Response) {
		var result map[string]interface{}

		err := json.Unmarshal(r.Body, &result)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}

		if sources, ok := result["sources"].([]interface{}); ok && len(sources) > 0 {
			if firstSource, ok := sources[0].(map[string]interface{}); ok {
				if file, ok := firstSource["file"].(string); ok {
					movieLink = file
				}
			}
		}
	})

	err := f.Collector.Visit(decryptUrl)
	if err != nil {
		fmt.Println("Error decrypting rabbit", err)
	}

	return movieLink
}

func (f *FlixHQProvider) KinoTime(url string) string {
	embeddedLink := f.GetEmbedLink(url)
	rabbitId := f.GetRabbitID(embeddedLink)
	kino := f.extractRabbitWasm(rabbitId)
	return kino
}
