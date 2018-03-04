package main

import (
	"encoding/json"
	"fmt"
	"github.com/mjibson/goon"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"net/http"
	"src/conf"
	"src/site"
)

var SubscribeUrl = "https://matopush.appspot.com/api/subscriber?"

func listHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	g := goon.NewGoon(r)

	query := datastore.NewQuery("SiteUpdateInfo").Filter("delete_flag=", false).Filter("public=", true)
	it := g.Run(query)

	var list []site.SiteUpdateInfo
	for {
		var s site.SiteUpdateInfo
		_, err := it.Next(&s)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "datastore get error.%v", err)
			break
		}
		list = append(list, s)
	}

	b, _ := json.Marshal(list)
	w.Write(b)
}

func addHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	g := goon.NewGoon(r)

	url := r.FormValue("siteUrl")
	endpoint := r.FormValue("endpoint")

	sui, isNewSite, err := site.Get(ctx, url)
	if err != nil {
		fmt.Fprint(w, "サイトの登録に失敗しました。")
		return
	}
	if isNewSite {
		g.Put(sui)
		if sui.HasHub {
			SubscribeRequest(ctx,
				SubscribeUrl+sui.SiteUrl,
				sui.FeedUrl,
				sui.HubUrl)
		}
	}
	fmt.Fprintf(w, "「%s」を追加しました。\n", sui.SiteTitle)
	err = conf.Update(ctx, endpoint, sui.FeedUrl, true)
	if err != nil {
		fmt.Fprint(w, "設定の更新に失敗しました。")
	} else {
		fmt.Fprintf(w, "サイト「%s」の更新を「通知する」に設定しました。", sui.SiteTitle)
	}
}
