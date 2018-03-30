package rss

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title    string   `xml:"title"`
	Item     []*Item  `xml:"item"`
	AtomLink AtomLink `xml:"link"`
}

type Item struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
	Description string `xml:"description"`
	PubDate string `xml:"pubDate"`
}

type AtomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}
