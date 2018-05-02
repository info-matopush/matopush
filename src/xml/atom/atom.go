package atom

import (
	"encoding/xml"
	"errors"
	"time"

	"github.com/info-matopush/matopush/src/content"
)

type atom struct {
	Title string  `xml:"title"`
	Entry []entry `xml:"entry"`
	Link  []link  `xml:"link"`
}

type entry struct {
	Title    string `xml:"title"`
	Link     []link `xml:"link"`
	Modified string `xml:"modified"`
	Updated  string `xml:"updated"`
	Summary  string `xml:"summary"`
}

type link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

func (e *entry) getContentURL() string {
	for _, link := range e.Link {
		if link.Rel == "alternate" {
			return link.Href
		}
	}
	return ""
}

// Analyze はXMLデータをFeed型へ変換する
// ATOMのスキーマについては下記を参照
// https://qiita.com/you88/items/e903fd463cf770688e1e
func Analyze(bytes []byte) (content.Feed, error) {
	feed := content.Feed{Type: "ATOM"}
	atom := atom{}
	err := xml.Unmarshal(bytes, &atom)
	if err != nil {
		return feed, err
	}

	feed.SiteTitle = atom.Title

	for _, item := range atom.Entry {
		cff := content.FromFeed{
			URL:     item.getContentURL(),
			Title:   item.Title,
			Summary: item.Summary,
		}

		// 2018-03-31T10:02:32Z MST
		layout1 := "2006-01-02T15:04:05Z MST"
		layout2 := "2006-01-02T15:04:05-07:00"
		cff.ModifyDate, err = time.Parse(layout1, item.Modified+" JST")
		if err != nil {
			cff.ModifyDate, err = time.Parse(layout2, item.Updated)
			// エラーは無視する
		}

		feed.Contents = append(feed.Contents, cff)
	}

	for _, l := range atom.Link {
		if l.Rel == "hub" {
			feed.HubURL = l.Href
		} else if l.Rel == "alternate" {
			feed.SiteURL = l.Href
		}
	}

	if len(feed.Contents) == 0 {
		return feed, errors.New("Can't find contents")
	}
	return feed, nil
}
