package doc

import (
	"encoding/xml"
	"net/http"

	"github.com/info-matopush/matopush/site"
	"github.com/info-matopush/matopush/xml/sitemap"
	"google.golang.org/appengine"
)

// SitemapHandler はsitemap.xmlを生成する
func SitemapHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	list := site.PublicList(ctx)

	w.Header().Set("content-type", "application/xml")
	w.Write([]byte(xml.Header))
	s := sitemap.FromLocation("https://matopush.appspot.com/")

	for _, site := range list {
		u := "https://matopush.appspot.com/invite?FeedURL=" + site.FeedURL
		s.Append(sitemap.CreateURLFromLocation(u))
	}

	buf, _ := xml.MarshalIndent(s, "", "  ")
	w.Write(buf)
}
