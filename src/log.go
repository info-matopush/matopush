package src

import (
	"encoding/base64"
	"hash/fnv"
	"net/http"
	"time"

	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type logInfo struct {
	Key        string    `datastore:"-" goon:"id"`
	Endpoint   string    `datastore:"endpoint,noindex"`
	SiteURL    string    `datastore:"site_url,noindex"`
	ContentURL string    `datastore:"content_url,noindex"`
	PushDate   time.Time `datastore:"push_date"`
	ReachDate  time.Time `datastore:"reach_date,noindex"`
	ClickDate  time.Time `datastore:"click_date,noindex"`
}

// LogCleanup は古いログを削除する
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

func endpointToKeyString(endpoint, contentURL string) string {
	h := fnv.New64a()
	h.Write([]byte(endpoint + ":" + contentURL))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// LogPush はPushしたログを記録する
func LogPush(ctx context.Context, endpoint, siteURL, contentURL string) {
	if true {
		return
	}

	g := goon.FromContext(ctx)
	l := &logInfo{
		Key:        endpointToKeyString(endpoint, contentURL),
		Endpoint:   endpoint,
		SiteURL:    siteURL,
		ContentURL: contentURL,
		PushDate:   time.Now(),
	}
	_, err := g.Put(l)
	if err != nil {
		log.Infof(ctx, "log write error. %v", err)
	}
}

// LogReach はEndpointに到達したことをログする
func LogReach(ctx context.Context, endpoint, contentURL string) {
	g := goon.FromContext(ctx)
	l := &logInfo{Key: endpointToKeyString(endpoint, contentURL)}
	err := g.Get(l)
	if err == nil {
		l.ReachDate = time.Now()
		g.Put(l)
	}
}

// LogClick はEndpoint(Notification)でクリックされたことをログする
func LogClick(ctx context.Context, endpoint, contentURL string) {
	g := goon.FromContext(ctx)
	l := &logInfo{Key: endpointToKeyString(endpoint, contentURL)}
	err := g.Get(l)
	if err == nil {
		l.ClickDate = time.Now()
		g.Put(l)
	}
}

// LogHandler はEndpointへの通知結果及び通知起因のユーザ操作をログする
func LogHandler(_ http.ResponseWriter, r *http.Request) {
	if true {
		return
	}
	ctx := appengine.NewContext(r)

	endpoint := r.FormValue("endpoint")
	url := r.FormValue("url")
	command := r.FormValue("command")

	if command == "click" {
		LogClick(ctx, endpoint, url)
	} else if command == "reach" {
		LogReach(ctx, endpoint, url)
	}
}
