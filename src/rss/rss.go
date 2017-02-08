package rss

type RSS struct {
	Channel Channel `xml:"channel"`
}

type Channel struct {
	Title string   `xml:"title"`
	Item  []*Item  `xml:"item"`
}

type Item struct {
	Title string  `xml:"title"`
	Link  string  `xml:"link"`
}

