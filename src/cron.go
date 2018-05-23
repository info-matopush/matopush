package src

import (
	"net/http"

	"github.com/info-matopush/matopush/src/conf"
	"github.com/info-matopush/matopush/src/endpoint"
	"github.com/info-matopush/matopush/src/site"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// CleanupHandler はDatastore上の不要データを削除する
func CleanupHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// 削除済み通知先のリストを取得する
	list := endpoint.GetAllDeleted(ctx)

	for _, ei := range list {
		err := conf.Cleanup(ctx, ei.Endpoint)
		if err == nil {
			ei.Cleanup(ctx)
		} else {
			log.Warningf(ctx, "cleanup error. %v", err)
		}
	}
	log.Infof(ctx, "cleanupしたendpoint数. %d", len(list))

	// 更新のないサイト情報を削除する
	err := site.DeleteUnnecessarySite(ctx)
	if err != nil {
		log.Errorf(ctx, "DeleteUnnecessarySite: %v", err)
	}

	// TODO: サイト情報と紐付かない購読情報を削除する

	// 古いログを削除する
	LogCleanup(ctx)
}

// HealthHandler は非表示のPushを全てのEndpointに送信し、
// 無効なEndpointを検出する
func HealthHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sui := site.UpdateInfo{
		Site: site.Site{
			SiteTitle: "まとプ",
			LatestContent: site.Content{
				Title: "",
			},
		},
	}

	sendPushAll(ctx, &sui)

	// 登録されているendpointの数を求める
	log.Infof(ctx, "有効なendpoint数. %d", endpoint.Count(ctx))
}

// SendNotificationHandler はサイトの更新をPushで通知する
func SendNotificationHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params := r.URL.Query()
	feedURL := params.Get("FeedURL")
	if feedURL == "" {
		log.Errorf(ctx, "FeedURL is empty")
		return
	}
	log.Infof(ctx, "FeedURL:%v", feedURL)

	ui, _, err := site.FromURL(ctx, feedURL)
	if err != nil {
		log.Errorf(ctx, "FromURL error %v", err)
		return
	}

	err = ui.CheckSite(ctx)
	if err != nil {
		log.Errorf(ctx, "CheckSite error %v", err)
		return
	}

	// 更新があればPushを行う
	sendPushWhenSiteUpdate(ctx, ui)

	// 更新された情報を保存する
	ui.Update(ctx)
}

// CronHandler は全てのサイトのFeedを読み直し、
// 更新があればPush送信を行う
func CronHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	siteList, err := site.List(ctx)
	if err != nil {
		log.Errorf(ctx, "get site failed. %v", err)
		return
	}

	for _, ui := range siteList {
		err := ui.CheckSite(ctx)
		if err != nil {
			// Feedの読み込みに失敗
			log.Warningf(ctx, "feedの読み込みに失敗 url:%s", ui.FeedURL)
			return
		}
		sendPushWhenSiteUpdate(ctx, &ui)
		ui.Update(ctx)
	}

	log.Infof(ctx, "site num:%d", len(siteList))
}

// RequestSubscribeHandler はHubURLを持つサイトに対し購読を要求する
func RequestSubscribeHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	siteList, err := site.List(ctx)
	if err != nil {
		log.Errorf(ctx, "get site failed. %v", err)
		return
	}

	for _, ui := range siteList {
		if ui.HubURL != "" {
			SubscribeRequest(ctx, SubscribeURL+ui.FeedURL, ui.FeedURL, ui.HubURL, ui.Secret)
		}
	}
}
