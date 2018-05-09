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

// ConfRemoveHandler は購読情報を削除する
func ConfRemoveHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	endpoint := r.FormValue("endpoint")
	feedURL := r.FormValue("feedUrl")

	conf.Delete(ctx, endpoint, feedURL)
}

// ConfListHandler は購読しているサイト購読情報を返却する
func ConfListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	endpoint := r.FormValue("endpoint")
	cList := conf.GetAllFromEndpoint(ctx, endpoint)

	var sList []site.UpdateInfo
	for _, c := range cList {
		sui, _, err := site.FromURL(ctx, c.FeedURL)
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

// ConfSiteHandler はサイト購読情報の更新を処理する
func ConfSiteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	endpoint := r.FormValue("endpoint")
	siteURL := r.FormValue("siteUrl")
	value := r.FormValue("value")

	enabled := true
	if value == "false" {
		enabled = false
	}
	sui, _, err := site.FromURL(ctx, siteURL)
	if err == nil {
		err := conf.Update(ctx, endpoint, sui.FeedURL, enabled)
		if err == nil {
			siteTitle := sui.SiteTitle
			if enabled {
				fmt.Fprintf(w, "サイト「%s」の更新を「通知する」に設定しました。", siteTitle)
			} else {
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
