package content

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

type Html struct {
	FeedURL  string
	ImageURL string
}

func HtmlParse(ctx context.Context, url string) (*Html, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	h := Html{}
	doc.Find("link").Each(func(i int, s *goquery.Selection) {
		rel, _ := s.Attr("rel")
		ref, _ := s.Attr("href")
		typ, _ := s.Attr("type")
		if rel == "alternate" {
			if typ == "application/rss+xml" {
				h.FeedURL = ref
			} else if typ == "application/atom+xml" {
				h.FeedURL = ref
			}
		}
	})
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		pro, _ := s.Attr("property")
		con, _ := s.Attr("content")
		if pro == "og:image" {
			h.ImageURL = con
		}
	})
	return &h, nil
}
