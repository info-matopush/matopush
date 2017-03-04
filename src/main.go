package main

import (
	"net/http"
)

func init() {
	http.HandleFunc("/api/regist", registHandler)
	http.HandleFunc("/api/unregist", unregistHandler)
	http.HandleFunc("/api/add", addHandler)
	http.HandleFunc("/api/key", keyHandler)
	http.HandleFunc("/api/dummy", dummyHander)
	http.HandleFunc("/api/cron", cronHandler)
	http.HandleFunc("/api/list", listHandler)
	http.HandleFunc("/api/cleanup", cleanupHandler)
	http.HandleFunc("/api/test", testHandler)
	http.HandleFunc("/api/conf/list", confListHandler)
	http.HandleFunc("/api/conf/site", confSiteHandler)
	http.HandleFunc("/api/mente", menteHandler)
	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/api/log", logHandler)
}
