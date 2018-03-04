package main

import (
	"encoding/json"
	"fmt"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
	"src/conf"
	"src/site"
)

func confListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	endpoint := r.FormValue("endpoint")
	cList := conf.ListFromEndpoint(ctx, endpoint)

	var sList []site.SiteUpdateInfo
	for _, c := range cList {
		sui, _, err := site.Get(ctx, c.FeedUrl)
		if err != nil {
			continue
		}
		if c.Enabled {
			sui.Value = "true"
		} else {
			sui.Value = "false"
		}
		sList = append(sList, *sui)
	}

	b, _ := json.Marshal(sList)
	w.Write(b)
}

func confSiteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	endpoint := r.FormValue("endpoint")
	siteUrl := r.FormValue("siteUrl")
	value := r.FormValue("value")

	enabled := true
	if value == "false" {
		enabled = false
	}
	sui, _, err := site.Get(ctx, siteUrl)
	if err == nil {
		err := conf.Update(appengine.NewContext(r), endpoint, sui.FeedUrl, enabled)
		if err == nil {
			siteTitle := sui.SiteTitle
			if value == "true" {
				fmt.Fprintf(w, "サイト「%s」の更新を「通知する」に設定しました。", siteTitle)
			} else if value == "false" {
				fmt.Fprintf(w, "サイト「%s」の更新を「通知しない」に設定しました。", siteTitle)
			}
			return
		} else {
			log.Infof(ctx, "conf.Updateに失敗 %v", err)
		}
	} else {
		log.Infof(ctx, "site.Getに失敗 %s, %v", siteUrl, err)
	}
	fmt.Fprint(w, "設定の更新に失敗しました。")
}
