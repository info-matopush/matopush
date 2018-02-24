package main

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
	"src/conf"
	"src/endpoint"
	"src/site"
)

type ContentInfo struct {
	ContentUrl   string
	ContentTitle string
}

type SiteInfo struct {
	SiteUrl   string
	SiteTitle string
}

func cleanupHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	// 削除済み通知先のリストを取得する
	list := endpoint.GetAllDeletedEndpoint(ctx)

	for _, ei := range list {
		err := conf.Cleanup(ctx, ei.Endpoint)
		if err == nil {
			endpoint.Cleanup(ctx, ei.Endpoint)
		} else {
			log.Warningf(ctx, "cleanup error. %v", err)
		}
	}
	log.Infof(ctx, "cleanupしたendpoint数. %d", len(list))

	// 古いログを削除する
	LogCleanup(ctx)
}

func healthHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sui := site.SiteUpdateInfo{SiteTitle: "まとプ", ContentTitle: ""}

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
		err := site.CheckSite(ctx, &sui)
		if err != nil {
			// Feedの読み込みに失敗
			// todo: どうしよう。とりあえず更新なしとする
			log.Warningf(ctx, "feedの読み込みに失敗 url:%s", sui.SiteUrl)
			return
		}
		sendPushWhenSiteUpdate(ctx, &sui)
	}

	log.Infof(ctx, "site num:%d", len(siteList))
}
