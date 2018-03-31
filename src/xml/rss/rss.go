package rss

import "github.com/info-matopush/matopush/src/content"

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title string `xml:"title"`
	Link  []Link `xml:"link"`
	Item  []Item `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Data string `xml:",chardata"`
}

func (r *RSS) ListContentFromFeed() []content.ContentFromFeed {
	var cff []content.ContentFromFeed
	for count, item := range r.Channel.Item {
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
