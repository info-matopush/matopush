package rdf

type RDF struct {
	Channel Channel `xml:"channel"`
	Item    []*Item `xml:"item"`
}

type Channel struct {
	Title    string   `xml:"title"`
	AtomLink AtomLink `xml:"link"`
}

type Item struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	Date        string `xml:"date"`
}

type AtomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}
