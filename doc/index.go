package doc

import (
	"html/template"
	"net/http"

	"github.com/info-matopush/matopush/site"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// IndexHandler はペインページを作成する
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	list := site.PublicList(ctx)

	// 雛形を読み込む
	t, err := template.ParseFiles("resources/template/index.html")
	if err != nil {
		log.Warningf(ctx, "template.ParseFiles err=%v", err)
		redirectErrorHTML(w)
		return
	}

	// html生成
	err = t.Execute(w, list)
	if err != nil {
		log.Warningf(ctx, "t.Execute err=%v", err)
		redirectErrorHTML(w)
		return
	}
}

func redirectErrorHTML(w http.ResponseWriter) {
	w.Header().Set("Location", "/")
	w.WriteHeader(http.StatusFound)
}
