package html

type Html struct {
	Head Head `xml:"head"`
}

type Head struct {
	Link []*Link `xml:"link"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
	Href string `xml:"href,attr"`
}
