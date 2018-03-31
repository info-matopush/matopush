package src

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/info-matopush/matopush/src/conf"
	"github.com/info-matopush/matopush/src/site"
	"google.golang.org/appengine"
)

var SubscribeURL = "https://matopush.appspot.com/api/subscriber?site="

func ListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	list, err := site.PublicList(ctx)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b, _ := json.Marshal(list)
	w.Write(b)
}

func AddHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	url := r.FormValue("siteUrl")
	endpoint := r.FormValue("endpoint")

	ui, isNewSite, err := site.FromUrl(ctx, url)
	if err != nil {
		fmt.Fprint(w, "サイトの登録に失敗しました。")
		return
	}
	if isNewSite {
		if ui.HubURL != "" {
			SubscribeRequest(ctx,
				SubscribeURL+ui.FeedURL,
				ui.FeedURL,
				ui.HubURL,
				ui.Secret)
		}
	}
	fmt.Fprintf(w, "「%s」を追加しました。\n", ui.SiteTitle)
	err = conf.Update(ctx, endpoint, ui.FeedURL, true)
	if err != nil {
		fmt.Fprint(w, "設定の更新に失敗しました。")
	} else {
		fmt.Fprintf(w, "サイト「%s」の更新を「通知する」に設定しました。", ui.SiteTitle)
	}
}
