package main

import (
	"net/http"

	"github.com/info-matopush/matopush/src"
	"github.com/info-matopush/matopush/src/conf"
	"github.com/info-matopush/matopush/src/cron"
	"github.com/info-matopush/matopush/src/site"
	"github.com/info-matopush/matopush/src/trace"
	"github.com/info-matopush/matopush/src/utility"
	"github.com/info-matopush/matopush/src/webpush"
	"google.golang.org/appengine"
)

func main() {
	// 招待
	http.HandleFunc("/invite", src.InviteHandler)
	// 登録・解除
	http.HandleFunc("/api/regist", webpush.RegistHandler)
	http.HandleFunc("/api/unregist", webpush.UnregistHandler)
	// ユーザ設定
	http.HandleFunc("/api/conf/site", conf.SiteHandler)
	http.HandleFunc("/api/conf/list", conf.ListHandler)
	http.HandleFunc("/api/conf/remove", conf.RemoveHandler)
	// サービス補助
	http.HandleFunc("/api/key", utility.KeyHandler)
	http.HandleFunc("/api/list", site.ListHandler)
	http.HandleFunc("/api/test", webpush.TestHandler)
	http.HandleFunc("/api/search", src.SearchHandler)
	http.HandleFunc("/api/subscriber", src.SubscriberHandler)
	http.HandleFunc("/api/log", trace.LogHandler)
	http.HandleFunc("/api/tunnel", src.TunnelHandler)
	// cron起動
	http.HandleFunc("/admin/api/cron", cron.CronHandler)
	http.HandleFunc("/admin/api/health", webpush.HealthHandler)
	http.HandleFunc("/admin/api/cleanup", cron.CleanupHandler)
	// taskqueue起動
	http.HandleFunc("/admin/api/publish", webpush.SendNotificationHandler)
	http.HandleFunc("/admin/api/request/subscribe", src.RequestSubscribeHandler)
	// メンテナンス
	http.HandleFunc("/admin/api/mainte", src.MainteHandler)
	http.HandleFunc("/admin/api/dummy", src.DummyHandler)
	appengine.Main()
}
