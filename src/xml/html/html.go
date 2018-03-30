package html

/*
import (
	"golang.org/x/net/context"
)
*/

type Html struct {
	Head Head   `xml:"head"`
	Meta []Meta `xml:"meta"`
}

type Head struct {
	Link []*Link `xml:"link"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}

type Meta struct {
	Property string `xml:"property,attr"`
	Name     string `xml:"name,attr"`
	Content  string `xml:"content,attr"`
}

/*
type Html struct {
	FeedUrl string
	HubUrl string
}

func Get(ctx context.Context, url string)(*Html, error) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		return nil, err
	}
	doc.Find("meta[rel=""")

}
*/