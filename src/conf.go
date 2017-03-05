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

func confListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	g := goon.NewGoon(r)

	endpoint := r.FormValue("endpoint")

	query := datastore.NewQuery("SiteSubscribe").Filter("endpoint=", endpoint)
	it := g.Run(query)

	sList := []site.SiteUpdateInfo{}
	for {
		var ss conf.SiteSubscribe
		_, err := it.Next(&ss)
		if err == datastore.Done {
			break
		}
		if err != nil {
			log.Errorf(ctx, "datastore get error.%v", err)
			return
		}

		sui := site.SiteUpdateInfo{SiteUrl: ss.SiteUrl}
		err = g.Get(&sui)
		if err != nil {
			continue
		}
		// 現在通知対象かどうかを設定する
		sui.Value = ss.Value
		sList = append(sList, sui)
	}

	b, _ := json.Marshal(sList)
	w.Write(b)
}

func confSiteHandler(w http.ResponseWriter, r *http.Request) {
	endpoint := r.FormValue("endpoint")
	siteUrl := r.FormValue("siteUrl")
	value := r.FormValue("value")

	siteTitle, err := conf.Update(appengine.NewContext(r), endpoint, siteUrl, value)

	if err != nil {
		fmt.Fprint(w, "設定の更新に失敗しました。")
	} else {
		if value == "true" {
			fmt.Fprintf(w, "サイト「%s」の更新を「通知する」に設定しました。", siteTitle)
		} else if value == "false" {
			fmt.Fprintf(w, "サイト「%s」の更新を「通知しない」に設定しました。", siteTitle)
		}
	}
}
