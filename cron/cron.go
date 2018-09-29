package cron

import (
	"net/http"
	"sync"

	"github.com/info-matopush/matopush/conf"
	"github.com/info-matopush/matopush/endpoint"
	"github.com/info-matopush/matopush/site"
	"github.com/info-matopush/matopush/trace"
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
	trace.LogCleanup(ctx)
}

// SiteCruisingHandler は全てのサイトのFeedを読み直し、
// 更新があればPush送信を行う
func SiteCruisingHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	siteList, err := site.List(ctx)
	if err != nil {
		log.Errorf(ctx, "get site failed. %v", err)
		return
	}

	if len(siteList) > 0 {
		// 登録されている全サイト毎に、更新通知用のタスクをキューに積む
		var wg sync.WaitGroup
		for _, ui := range siteList {
			wg.Add(1)
			go func(ui site.UpdateInfo) {
				defer wg.Done()

				err = ui.CheckSite(ctx)
				if err != nil {
					log.Errorf(ctx, "CheckSite error %v", err)
					return
				}

				if ui.UpdateFlg {
					// チェック結果を保存
					ui.Update(ctx)

					// 更新がある場合はプッシュで通知する
					PutTaskSendNotifiation(ctx, ui.FeedURL)
				}
			}(ui)
		}
		wg.Wait()
	}
	log.Infof(ctx, "巡回したサイトの数:%d", len(siteList))
}
