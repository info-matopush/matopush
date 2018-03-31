package rdf

import "github.com/info-matopush/matopush/src/content"

type RDF struct {
	Channel Channel `xml:"channel"`
	Item    []Item  `xml:"item"`
}

type Channel struct {
	Title string `xml:"title"`
	Link  []Link `xml:"link"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Date        string `xml:"date"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Data string `xml:",chardata"`
}

func (r *RDF) ListContentFromFeed() []content.ContentFromFeed {
	var cff []content.ContentFromFeed
	for count, item := range r.Item {
		cff = append(cff, content.ContentFromFeed{
			URL:     item.Link,
			Title:   item.Title,
			Summary: item.Description,
		})
		if count > 5 {
			break
		}
	}
	return cff
}
