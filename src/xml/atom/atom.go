package atom

import (
	"github.com/info-matopush/matopush/src/content"
)

type Feed struct {
	Title string  `xml:"title"`
	Entry []Entry `xml:"entry"`
	Link  []Link  `xml:"link"`
}

type Entry struct {
	Title    string `xml:"title"`
	Link     []Link `xml:"link"`
	Modified string `xml:"modified"`
	Summary  string `xml:"summary"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

func (e *Entry) getContentUrl() string {
	for _, link := range e.Link {
		if link.Rel == "alternate" {
			return link.Href
		}
	}
	return ""
}

func (f *Feed) ListContentFromFeed() []content.ContentFromFeed {
	var cff []content.ContentFromFeed
	for count, entry := range f.Entry {
		u := entry.getContentUrl()
		if u != "" {
			cff = append(cff, content.ContentFromFeed{
				URL:     entry.getContentUrl(),
				Title:   entry.Title,
				Summary: entry.Summary,
			})
		}
		if count > 5 {
			break
		}
	}
	return cff
}
