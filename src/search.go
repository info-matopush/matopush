package src

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/info-matopush/matopush/src/content"
	"golang.org/x/oauth2/google"
	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type searchResult struct {
	Items []searchItem `json:"items"`
	Next  int          `json:"next"`
}

type searchItem struct {
	Title   string `json:"title"`
	Snippet string `json:"snippet"`
	FeedURL string `json:"feedURL"`
}

// SearchHandler はWeb検索を行い結果を返す
func SearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	p := NewFromContext(ctx)
	key := p.GetString("google.custom.search.apikey", "undefined")
	id := p.GetString("google.search.engine.id", "undefined")

	log.Infof(ctx, "google.search.engine.id: %v", id)

	_ = r.FormValue("endpoint")
	keyword := r.FormValue("keyword")
	pos, err := strconv.Atoi(r.FormValue("position"))
	if err != nil {
		pos = 1
	}

	conf, err := google.JWTConfigFromJSON([]byte(key), "https://www.googleapis.com/auth/cse")
	if err != nil {
		log.Infof(ctx, "JWTConfigFromJSON error. %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cseService, err := customsearch.New(conf.Client(ctx))
	if err != nil {
		log.Infof(ctx, "customsearch.New err. %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	search := cseService.Cse.List(keyword)
	search.Cx(id)
	search.Lr("lang_ja")
	search.Start(int64(pos))

	s, err := search.Do()
	if err != nil {
		log.Infof(ctx, "search.Do error. keyword %v, %v", keyword, err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var sr searchResult
	for _, i := range s.Items {
		h, err := content.ParseHTML(ctx, i.Link)
		if err == nil && h.FeedURL != "" {
			si := searchItem{
				i.Title,
				i.Snippet,
				h.FeedURL,
			}
			sr.Items = append(sr.Items, si)
		}
	}
	sr.Next = pos + 10

	b, err := json.Marshal(sr)
	if err != nil {
		log.Infof(ctx, "json.Marshal error. %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Write(b)
}
