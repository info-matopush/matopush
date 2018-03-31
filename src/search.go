package src

import (
	"encoding/json"
	"net/http"
	"strconv"

	"golang.org/x/oauth2/google"
	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var j = ``

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	p := NewFromContext(ctx)
	//key := p.GetString("google.custom.search.apikey", "undefined")
	id := p.GetString("google.search.engine.id", "undefined")

	log.Infof(ctx, "google.search.engine.id: %v", id)

	// endpoint := r.FormValue("endpoint")
	keyword := r.FormValue("keyword")
	pos, err := strconv.Atoi(r.FormValue("position"))
	if err != nil {
		pos = 1
	}

	conf, err := google.JWTConfigFromJSON([]byte(j), "https://www.googleapis.com/auth/cse")
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
	search.Start(int64(pos))

	s, err := search.Do()
	if err != nil {
		log.Infof(ctx, "search.Do error. keyword %v, %v", keyword, err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	b, err := json.Marshal(s)
	if err != nil {
		log.Infof(ctx, "json.Marshal error. %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Write(b)
}
