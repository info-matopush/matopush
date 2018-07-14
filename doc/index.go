package doc

import (
	"html/template"
	"net/http"

	"github.com/info-matopush/matopush/site"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

type updateInfoView struct {
	site.UpdateInfo
	HasHub       bool
	SafeSiteIcon string
	HasIcon      bool
}

// IndexHandler はペインページを作成する
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	list := site.PublicList(ctx)

	var vlist []updateInfoView
	for _, u := range list {
		v := updateInfoView{
			UpdateInfo:   u,
			HasHub:       ("" != u.HubURL),
			SafeSiteIcon: u.SiteIcon.TunneledURL(),
			HasIcon:      ("" != u.SiteIcon),
		}
		vlist = append(vlist, v)
	}

	// 雛形を読み込む
	t, err := template.ParseFiles("resources/template/index.html")
	if err != nil {
		log.Warningf(ctx, "template.ParseFiles err=%v", err)
		redirectErrorHTML(w)
		return
	}

	// html生成
	err = t.Execute(w, vlist)
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
