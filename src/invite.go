package src

import (
	"html/template"
	"net/http"

	"github.com/info-matopush/matopush/src/site"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// InviteHandler は招待用のリンクを処理する
func InviteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params := r.URL.Query()
	feedURL := params.Get("FeedURL")

	s, err := site.FromFeedURL(ctx, feedURL)
	if err != nil {
		log.Warningf(ctx, "site.FromFeedURL err=%v", err)
		redirectIndexHTML(w)
		return
	}
	t, err := template.ParseFiles("template/invite.html")
	if err != nil {
		log.Warningf(ctx, "template.ParseFileserr=%v", err)
		redirectIndexHTML(w)
		return
	}

	err = t.Execute(w, s)
	if err != nil {
		log.Warningf(ctx, "t.Execute err=%v", err)
		redirectIndexHTML(w)
		return
	}
}

func redirectIndexHTML(w http.ResponseWriter) {
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}
