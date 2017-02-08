package main

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
	"src/endpoint"
	"src/site"
)

type ContentInfo struct {
	ContentUrl   string
	ContentTitle string
}

type SiteInfo struct {
	SiteUrl      string
	SiteTitle    string     // todo: titleタグから取得するように後で変更する
}

func healthHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sui := site.SiteUpdateInfo{SiteTitle:"まとプ", ContentTitle:""}

	sendPushAll(ctx, &sui)

	// 登録されているendpointの数を求める
	log.Infof(ctx, "有効なendpoint数. %d", endpoint.Count(ctx))
}

func cronHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	var siteList []site.SiteUpdateInfo
	err := site.GetAll(ctx, &siteList)
	if err != nil {
		log.Errorf(ctx, "get site failed. %v", err)
	}

	for _, sui := range siteList {
		site.CheckSite(ctx, &sui)
		sendPushWhenSiteUpdate(ctx, &sui)
	}

	log.Infof(ctx, "site num:%d", len(siteList))
}

