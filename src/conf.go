package src

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/info-matopush/matopush/src/conf"
	"github.com/info-matopush/matopush/src/site"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func ConfListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	endpoint := r.FormValue("endpoint")
	cList := conf.ListFromEndpoint(ctx, endpoint)

	var sList []site.UpdateInfo
	for _, c := range cList {
		sui, _, err := site.FromUrl(ctx, c.FeedUrl)
		if err != nil {
			continue
		}
		sui.Value = c.Enabled
		sList = append(sList, *sui)
		log.Infof(ctx, "UpdateInfo %v", sui)
	}

	b, _ := json.Marshal(sList)
	w.Write(b)
}

func ConfSiteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	endpoint := r.FormValue("endpoint")
	siteURL := r.FormValue("siteUrl")
	value := r.FormValue("value")

	enabled := true
	if value == "false" {
		enabled = false
	}
	sui, _, err := site.FromUrl(ctx, siteURL)
	if err == nil {
		err := conf.Update(appengine.NewContext(r), endpoint, sui.FeedURL, enabled)
		if err == nil {
			siteTitle := sui.SiteTitle
			if value == "true" {
				fmt.Fprintf(w, "サイト「%s」の更新を「通知する」に設定しました。", siteTitle)
			} else if value == "false" {
				fmt.Fprintf(w, "サイト「%s」の更新を「通知しない」に設定しました。", siteTitle)
			}
		} else {
			log.Infof(ctx, "conf.Updateに失敗 %v", err)
			fmt.Fprint(w, "設定の更新に失敗しました。")
		}
	} else {
		log.Infof(ctx, "site.Getに失敗 %s, %v", siteURL, err)
		fmt.Fprint(w, "設定の更新に失敗しました。")
	}
}
