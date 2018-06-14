package src

import (
	"net/http"
	"sync"

	"github.com/info-matopush/matopush/src/conf"
	"github.com/info-matopush/matopush/src/endpoint"
	"github.com/info-matopush/matopush/src/site"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/taskqueue"
)

// PutTaskSendNotifiation はタスクキューにWebPush用のタスクを積む
func PutTaskSendNotifiation(ctx context.Context, feedURL string) {
	t := taskqueue.NewPOSTTask("/admin/api/publish?FeedURL="+feedURL, nil)
	if _, err := taskqueue.Add(ctx, t, "background-job"); err != nil {
		log.Errorf(ctx, "taskqueue.Add error %v", err)
	}
}

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

	// 登録されている全サイト毎に、更新通知用のタスクをキューに積む
	var wg sync.WaitGroup
	for _, ui := range siteList {
		wg.Add(1)
		go func(ui site.UpdateInfo) {
			defer wg.Done()
			PutTaskSendNotifiation(ctx, ui.FeedURL)
		}(ui)
	}
	wg.Done()
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
