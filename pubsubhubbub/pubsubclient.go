package pubsubhubbub

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/info-matopush/matopush/cron"
	"github.com/info-matopush/matopush/site"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

// SubscribeURL はHubからの通知を受信するURL
var SubscribeURL = "https://matopush.appspot.com/api/subscriber?site="

// SubscriberHandler はHubから通知された購読情報を処理する
func SubscriberHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if r.Method == "GET" {
		verify(w, r)
		return
	}

	// サイトが登録済みのものか？
	params := r.URL.Query()
	_, err := site.FromFeedURL(ctx, params.Get("site"))
	if err == nil {
		// サイト更新情報をWebPushで通知するためのタスクをキューに積む
		cron.PutTaskSendNotifiation(ctx, params.Get("site"))
	}
}

// https://www.w3.org/TR/websub
// 5.3.1 Verification Details
// The subscriber MUST confirm that the hub.topic corresponds to a pending subscription or
// unsubscription that it wishes to carry out. If so, the subscriber MUST respond with an HTTP success (2xx)
// code with a response body equal to the hub.challenge parameter.
func verify(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params := r.URL.Query()
	ui, isNewSite, err := site.FromURL(ctx, params.Get("site"))
	if err == nil && !isNewSite {
		// 購読対象URLの場合
		if params.Get("hub.verify_token") == ui.Secret {
			w.Write([]byte(params.Get("hub.challenge")))
			log.Infof(ctx, "pubsubhubbubからの通知を有効にしました site=%v", params.Get("site"))
			return
		}
	}
	// 購読対象外の場合はステータスコードを4xxにする
	w.WriteHeader(http.StatusNotFound)
}

// SubscribeRequest はHubに購読を要求する
func SubscribeRequest(ctx context.Context, callbackURL, topic, hub, secret string) {
	body := url.Values{}
	body.Set("hub.mode", "subscribe")
	body.Add("hub.topic", topic)
	body.Add("hub.callback", callbackURL)
	body.Add("hub.verify", "async")
	body.Add("hub.verify_token", secret)

	req, err := http.NewRequest("POST", hub, bytes.NewBufferString(body.Encode()))
	if err != nil {
		// error時は以降の処理をしない
		log.Infof(ctx, "SubscribeRequest err %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		log.Infof(ctx, "SubscribeRequest: err %v", err)
		return
	}
	defer resp.Body.Close()
	reason, _ := ioutil.ReadAll(resp.Body)
	log.Infof(ctx, "reason %v, topic %v", string(reason), topic)
}

// RequestSubscribeHandler はHubURLを持つサイトに対し購読を要求する
func RequestSubscribeHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	siteList, err := site.List(ctx)
	if err != nil {
		log.Errorf(ctx, "get site failed. %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, ui := range siteList {
		wg.Add(1)
		go func(ui site.UpdateInfo) {
			defer wg.Done()
			if ui.HubURL != "" {
				url := strings.Replace(SubscribeURL, "matopush", appengine.AppID(ctx), -1)
				SubscribeRequest(ctx, url+ui.FeedURL, ui.FeedURL, ui.HubURL, ui.Secret)
			}
		}(ui)
	}
	wg.Wait()
}
