package doc

import (
	"encoding/xml"
	"net/http"
	"strings"

	"github.com/info-matopush/matopush/site"
	"github.com/info-matopush/matopush/xml/sitemap"
	"google.golang.org/appengine"
)

// SitemapHandler はsitemap.xmlを生成する
func SitemapHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	mainURL := "https://matopush.appspot.com/"
	mainURL = strings.Replace(mainURL, "matopush", appengine.AppID(ctx), -1)

	list := site.PublicList(ctx)

	w.Header().Set("content-type", "application/xml")
	w.Write([]byte(xml.Header))
	s := sitemap.FromLocation(mainURL)

	for _, site := range list {
		req, _ := http.NewRequest("GET", mainURL+"invite", nil)
		q := req.URL.Query()
		q.Add("FeedURL", site.FeedURL)
		req.URL.RawQuery = q.Encode()
		s.Append(sitemap.CreateURLFromLocation(req.URL.String()))
	}

	buf, _ := xml.MarshalIndent(s, "", "  ")
	w.Write(buf)
}
