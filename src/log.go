package main

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
)

func logHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	// プッシュ通知を契機にウェブ遷移した。ログする。
	endpoint := r.FormValue("endpoint")
	url := r.FormValue("url")
	log.Infof(ctx, "プッシュ通知からWebページ閲覧実施した。%s, %s", url, endpoint)

	// todo: このデータをbigdataに記録する
}
