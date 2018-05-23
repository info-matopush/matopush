package main

import (
	"net/http"

	"github.com/info-matopush/matopush/src"
	"google.golang.org/appengine"
)

func main() {
	// 登録・解除
	http.HandleFunc("/api/regist", src.RegistHandler)
	http.HandleFunc("/api/unregist", src.UnregistHandler)
	// ユーザ設定
	http.HandleFunc("/api/conf/site", src.ConfSiteHandler)
	http.HandleFunc("/api/conf/list", src.ConfListHandler)
	http.HandleFunc("/api/conf/remove", src.ConfRemoveHandler)
	// サービス補助
	http.HandleFunc("/api/key", src.KeyHandler)
	http.HandleFunc("/api/list", src.ListHandler)
	http.HandleFunc("/api/test", src.TestHandler)
	http.HandleFunc("/api/search", src.SearchHandler)
	http.HandleFunc("/api/subscriber", src.SubscriberHandler)
	http.HandleFunc("/api/log", src.LogHandler)
	// cron起動
	http.HandleFunc("/admin/api/cron", src.CronHandler)
	http.HandleFunc("/admin/api/health", src.HealthHandler)
	http.HandleFunc("/admin/api/cleanup", src.CleanupHandler)
	// taskqueue起動
	http.HandleFunc("/admin/api/publish", src.SendNotificationHandler)
	http.HandleFunc("/admin/api/request/subscribe", src.RequestSubscribeHandler)
	// メンテナンス
	http.HandleFunc("/admin/api/mainte", src.MainteHandler)
	http.HandleFunc("/admin/api/dummy", src.DummyHandler)
	appengine.Main()
}
