package main

import (
	"net/http"

	"github.com/info-matopush/matopush/conf"
	"github.com/info-matopush/matopush/cron"
	"github.com/info-matopush/matopush/site"
	"github.com/info-matopush/matopush/trace"
	"github.com/info-matopush/matopush/utility"
	"github.com/info-matopush/matopush/webpush"
	"google.golang.org/appengine"
)

func main() {
	// 招待
	http.HandleFunc("/invite", InviteHandler)
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
	http.HandleFunc("/api/search", utility.SearchHandler)
	http.HandleFunc("/api/subscriber", utility.SubscriberHandler)
	http.HandleFunc("/api/log", trace.LogHandler)
	http.HandleFunc("/api/tunnel", utility.TunnelHandler)
	// cron起動
	http.HandleFunc("/admin/api/cron", cron.CronHandler)
	http.HandleFunc("/admin/api/health", webpush.HealthHandler)
	http.HandleFunc("/admin/api/cleanup", cron.CleanupHandler)
	// taskqueue起動
	http.HandleFunc("/admin/api/publish", webpush.SendNotificationHandler)
	http.HandleFunc("/admin/api/request/subscribe", utility.RequestSubscribeHandler)
	// メンテナンス
	http.HandleFunc("/admin/api/mainte", MainteHandler)
	http.HandleFunc("/admin/api/dummy", DummyHandler)
	appengine.Main()
}
