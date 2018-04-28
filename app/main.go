package main

import (
	"net/http"

	"github.com/info-matopush/matopush/src"
	"google.golang.org/appengine"
)

func main() {
	http.HandleFunc("/api/regist", src.RegistHandler)
	http.HandleFunc("/api/unregist", src.UnregistHandler)
	http.HandleFunc("/api/add", src.AddHandler)
	http.HandleFunc("/api/key", src.KeyHandler)
	http.HandleFunc("/api/dummy", src.DummyHandler)
	http.HandleFunc("/api/cron", src.CronHandler)
	http.HandleFunc("/api/list", src.ListHandler)
	http.HandleFunc("/api/cleanup", src.CleanupHandler)
	http.HandleFunc("/api/test", src.TestHandler)
	http.HandleFunc("/api/conf/list", src.ConfListHandler)
	http.HandleFunc("/api/conf/site", src.ConfSiteHandler)
	http.HandleFunc("/api/conf/remove", src.ConfRemoveHandler)
	http.HandleFunc("/api/mainte", src.MainteHandler)
	http.HandleFunc("/api/health", src.HealthHandler)
	http.HandleFunc("/api/log", src.LogHandler)
	http.HandleFunc("/api/subscriber", src.SubscriberHandler)
	http.HandleFunc("/api/search", src.SearchHandler)
	appengine.Main()
}
