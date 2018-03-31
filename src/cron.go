package src

import (
	"net/http"

	"github.com/info-matopush/matopush/src/conf"
	"github.com/info-matopush/matopush/src/endpoint"
	"github.com/info-matopush/matopush/src/site"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

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

	// todo: サイト情報と紐付かない購読情報を削除する

	// 古いログを削除する
	LogCleanup(ctx)
}

func HealthHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sui := site.UpdateInfo{SiteTitle: "まとプ", ContentTitle: ""}

	sendPushAll(ctx, &sui)

	// 登録されているendpointの数を求める
	log.Infof(ctx, "有効なendpoint数. %d", endpoint.Count(ctx))
}

func CronHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	siteList, err := site.List(ctx)
	if err != nil {
		log.Errorf(ctx, "get site failed. %v", err)
	}

	for _, ui := range siteList {
		err := ui.CheckSite(ctx)
		if err != nil {
			// Feedの読み込みに失敗
			log.Warningf(ctx, "feedの読み込みに失敗 url:%s", ui.FeedURL)
			return
		}
		sendPushWhenSiteUpdate(ctx, &ui)

		// huburlが設定されていた場合は積極的に利用する
		if ui.UpdateFlg && ui.HubURL != "" {
			log.Infof(ctx, "use pubsub %v", ui.FeedURL)
			SubscribeRequest(ctx, SubscribeURL+ui.FeedURL, ui.FeedURL, ui.HubURL, ui.Secret)
		}
		ui.Update(ctx)
	}

	log.Infof(ctx, "site num:%d", len(siteList))
}
