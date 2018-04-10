package rdf

import (
	"encoding/xml"
	"errors"
	"time"

	"github.com/info-matopush/matopush/src/content"
)

type rdf struct {
	Channel channel `xml:"channel"`
	Item    []item  `xml:"item"`
}

type channel struct {
	Title string `xml:"title"`
	Link  []link `xml:"link"`
}

type item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Date        string `xml:"date"`
}

type link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
	Data string `xml:",chardata"`
}

// Analyze はXMLデータをFeed型へ変換する
// RDFのスキーマについては下記を参照
// https://qiita.com/you88/items/e903fd463cf770688e1e
func Analyze(bytes []byte) (content.Feed, error) {
	feed := content.Feed{Type: "RSS 1.0"}
	rdf := rdf{}
	err := xml.Unmarshal(bytes, &rdf)
	if err != nil {
		return feed, err
	}

	feed.SiteTitle = rdf.Channel.Title

	for _, item := range rdf.Item {
		cff := content.FromFeed{
			URL:     item.Link,
			Title:   item.Title,
			Summary: item.Description,
		}

		// 2018-03-31T10:02:32+09:00
		layout1 := "2006-01-02T15:04:05-07:00"
		cff.ModifyDate, err = time.Parse(layout1, item.Date)
		// エラーは無視する

		feed.Contents = append(feed.Contents, cff)
	}

	for _, l := range rdf.Channel.Link {
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
