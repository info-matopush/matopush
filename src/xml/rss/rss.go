package rss

import (
	"encoding/xml"
	"errors"
	"time"

	"github.com/info-matopush/matopush/src/content"
)

type rss struct {
	Channel channel `xml:"channel"`
}

type channel struct {
	Title string `xml:"title"`
	Link  []link `xml:"link"`
	Item  []item `xml:"item"`
}

type item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

type link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Data string `xml:",chardata"`
}

// Analyze はXMLデータをFeed型へ変換する
// RSS 2.0のスキーマについては下記を参照
// http://www.futomi.com/lecture/japanese/rss20.html
func Analyze(bytes []byte) (content.Feed, error) {
	feed := content.Feed{Type: "RSS 2.0"}
	rss := rss{}
	err := xml.Unmarshal(bytes, &rss)
	if err != nil {
		return feed, err
	}

	feed.SiteTitle = rss.Channel.Title

	for _, item := range rss.Channel.Item {
		cff := content.FromFeed{
			URL:     item.Link,
			Title:   item.Title,
			Summary: item.Description,
		}

		layout1 := "Mon, 02 Jan 2006 15:04:05 -0700"
		layout2 := "Mon, 02 Jan 2006 15:04:05 MST"
		cff.ModifyDate, err = time.Parse(layout1, item.PubDate)
		if err != nil {
			cff.ModifyDate, err = time.Parse(layout2, item.PubDate)
			// エラーは無視する
		}

		feed.Contents = append(feed.Contents, cff)
	}

	for _, l := range rss.Channel.Link {
		if l.Rel == "hub" {
			feed.HubURL = l.Href
		} else if l.Data != "" {
			feed.SiteURL = l.Data
		}
	}

	if len(feed.Contents) == 0 {
		return feed, errors.New("Can't find contents")
	}
	return feed, nil
}
