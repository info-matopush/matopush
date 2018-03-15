package main

import (
	"encoding/json"
	"fmt"
	"github.com/mjibson/goon"
	"google.golang.org/appengine"
	"net/http"
	"src/conf"
	"src/site"
)

var SubscribeUrl = "https://matopush.appspot.com/api/subscriber?site="

func listHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	list, err := site.PublicList(ctx)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	b, _ := json.Marshal(list)
	w.Write(b)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	g := goon.NewGoon(r)

	url := r.FormValue("siteUrl")
	endpoint := r.FormValue("endpoint")

	ui, isNewSite, err := site.FromUrl(ctx, url)
	if err != nil {
		fmt.Fprint(w, "サイトの登録に失敗しました。")
		return
	}
	if isNewSite {
		g.Put(ui)
		if ui.HubUrl != "" {
			SubscribeRequest(ctx,
				SubscribeUrl+ui.FeedUrl,
				ui.FeedUrl,
				ui.HubUrl,
				ui.Secret)
		}
	}
	fmt.Fprintf(w, "「%s」を追加しました。\n", ui.SiteTitle)
	err = conf.Update(ctx, endpoint, ui.FeedUrl, true)
	if err != nil {
		fmt.Fprint(w, "設定の更新に失敗しました。")
	} else {
		fmt.Fprintf(w, "サイト「%s」の更新を「通知する」に設定しました。", ui.SiteTitle)
	}
}
