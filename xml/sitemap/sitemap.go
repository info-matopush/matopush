package sitemap

import (
	"encoding/xml"
	"time"
)

// Sitemap はサイト全体のページに関する情報
type Sitemap struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URL     []URL    `xml:"url"`
}

// URL はサイト内のページに関する情報
type URL struct {
	Location       string `xml:"loc"`
	LastModifyDate string `xml:"lastmod"`
	ChangeFreq     string `xml:"changefreq"`
	Priority       string `xml:"priority"`
}

// FromURL はURLからSitemapを作成する
func FromURL(url URL) Sitemap {
	var u = []URL{url}
	return Sitemap{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URL:   u,
	}
}

// FromLocation はlocationからSitemapを作成する
func FromLocation(location string) Sitemap {
	u := CreateURLFromLocation(location)
	return FromURL(u)
}

// Append はURLを追加する
func (s *Sitemap) Append(url URL) {
	s.URL = append(s.URL, url)
}

// CreateURL はURLを作成する
func CreateURL(location, changeFreq string, lastModifyDate time.Time) URL {
	layout := "2006-01-02"
	m := lastModifyDate.Format(layout)
	return URL{
		Location:       location,
		LastModifyDate: m,
		ChangeFreq:     changeFreq,
		Priority:       "1.0",
	}
}

// CreateURLFromLocation はURLを作成する
func CreateURLFromLocation(location string) URL {
	return CreateURL(location, "weekly", time.Now())
}
