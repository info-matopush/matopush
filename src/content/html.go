package content

import (
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/context"
	"google.golang.org/appengine/urlfetch"
)

// HTML はhtmlから取得できる情報を持つ
type HTML struct {
	FeedURL  string
	ImageURL string
	IconURL  string
}

// HTMLParse はurlで取得したHTMLを解析して
// 中に含まれる情報を返す
func HTMLParse(ctx context.Context, url string) (*HTML, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return nil, err
	}

	h := HTML{}
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
		} else if rel == "icon" {
			h.IconURL = ref
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
