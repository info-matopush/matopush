package rdf

type RDF struct {
	Channel Channel `xml:"channel"`
	Item    []*Item `xml:"item"`
}

type Channel struct {
	Title string `xml:"title"`
}

type Item struct {
	Title string `xml:"title"`
	Link  string `xml:"link"`
}
