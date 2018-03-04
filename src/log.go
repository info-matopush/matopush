package main

import (
	"encoding/base64"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"hash/fnv"
	"net/http"
	"time"
)

type logInfo struct {
	Key        string    `datastore:"-" goon:"id"`
	Endpoint   string    `datastore:"endpoint,noindex"`
	SiteUrl    string    `datastore:"site_url,noindex"`
	ContentUrl string    `datastore:"content_url,noindex"`
	PushDate   time.Time `datastore:"push_date"`
	ReachDate  time.Time `datastore:"reach_date,noindex"`
	ClickDate  time.Time `datastore:"click_date,noindex"`
}

func LogCleanup(ctx context.Context) {
	limit := time.Now().Add(time.Duration(-7*24) * time.Hour)
	query := datastore.NewQuery("logInfo").Filter("push_date<", limit).KeysOnly()
	g := goon.FromContext(ctx)
	keys, err := g.GetAll(query, nil)
	if err != nil {
		log.Errorf(ctx, "log cleanup error(GetAll). %v", err)
		return
	}
	err = g.DeleteMulti(keys)
	if err != nil {
		log.Errorf(ctx, "log cleanup error(DeleteMulti. %v", err)
		return
	}
}

func endpointToKeyString(endpoint, contentUrl string) string {
	h := fnv.New64a()
	h.Write([]byte(endpoint + ":" + contentUrl))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func LogPush(ctx context.Context, endpoint, siteUrl, contentUrl string) {
	g := goon.FromContext(ctx)
	l := &logInfo{
		Key:        endpointToKeyString(endpoint, contentUrl),
		Endpoint:   endpoint,
		SiteUrl:    siteUrl,
		ContentUrl: contentUrl,
		PushDate:   time.Now(),
	}
	_, err := g.Put(l)
	if err != nil {
		log.Infof(ctx, "log write error. %v", err)
	}
}

func LogReach(ctx context.Context, endpoint, contentUrl string) {
	g := goon.FromContext(ctx)
	l := &logInfo{Key: endpointToKeyString(endpoint, contentUrl)}
	err := g.Get(l)
	if err == nil {
		l.ReachDate = time.Now()
		g.Put(l)
	}
}

func LogClick(ctx context.Context, endpoint, contentUrl string) {
	g := goon.FromContext(ctx)
	l := &logInfo{Key: endpointToKeyString(endpoint, contentUrl)}
	err := g.Get(l)
	if err == nil {
		l.ClickDate = time.Now()
		g.Put(l)
	}
}

func logHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	// プッシュ通知を契機にウェブ遷移した。ログする。
	endpoint := r.FormValue("endpoint")
	url := r.FormValue("url")
	command := r.FormValue("command")
	log.Infof(ctx, "プッシュ通知からWebページ閲覧した。%s, %s", url, endpoint)

	if command == "click" {
		//		LogClick(ctx, endpoint, url)
	} else if command == "reach" {
		//		LogReach(ctx, endpoint, url)
	}
}
