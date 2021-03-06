package utility

import (
	"strings"

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

// ParseHTML はurlで取得したHTMLを解析して
// 中に含まれる情報を返す
func ParseHTML(ctx context.Context, url string) (HTML, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		return HTML{}, err
	}

	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return HTML{}, err
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
			if !strings.HasPrefix(ref, "http") {
				ref = resp.Request.URL.Scheme +
					"://" + resp.Request.Host + ref
			}
			h.IconURL = ref
		}
	})
	doc.Find("meta").Each(func(i int, s *goquery.Selection) {
		pro, _ := s.Attr("property")
		nam, _ := s.Attr("name")
		con, _ := s.Attr("content")
		if pro == "og:image" {
			h.ImageURL = con
		} else if nam == "og:image" {
			// サイトによってはproperty="og:image"ではなくname="og:image"の場合がある
			h.ImageURL = con
		}
	})
	return h, nil
}
