package atom

type Feed struct {
	Title string   `xml:"title"`
	Entry []*Entry `xml:"entry"`
}

type Entry struct {
	Title string    `xml:"title"`
	Link  []*Link   `xml:"link"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}
