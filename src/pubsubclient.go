package main

import (
	"bytes"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/http"
	"net/url"
	"src/site"
)

func subscriberHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	if r.Method == "GET" {
		verify(w, r)
		return
	}

	// POST時はfeedが送られてくる
	params := r.URL.Query()
	defer r.Body.Close()
	body, _ := ioutil.ReadAll(r.Body)

	// 購読情報があればpushを行う
	sui, _ := site.CheckSiteByFeed(ctx, params.Get("site"), body)
	if sui != nil {
		sendPushWhenSiteUpdate(ctx, sui)
		sui.Update(ctx)
		return
	}
	// 購読対象外の場合はステータスコードを4xxにする
	w.WriteHeader(http.StatusNotFound)
}

// https://www.w3.org/TR/websub
// 5.3.1 Verification Details
// The subscriber MUST confirm that the hub.topic corresponds to a pending subscription or
// unsubscription that it wishes to carry out. If so, the subscriber MUST respond with an HTTP success (2xx)
// code with a response body equal to the hub.challenge parameter.
func verify(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params := r.URL.Query()
	_, isNewSite, err := site.FromUrl(ctx, params.Get("site"))
	if err != nil && !isNewSite {
		// 購読対象URLの場合
		w.Write([]byte(params.Get("hub.challenge")))
		log.Infof(ctx, "pubsubhubbubからの通知を有効にしました site=%s", params.Get("site"))
		return
	}
	// 購読対象外の場合はステータスコードを4xxにする
	w.WriteHeader(http.StatusNotFound)
}

func SubscribeRequest(ctx context.Context, callbackUrl, topic, hub string) {
	body := url.Values{}
	body.Set("hub.mode", "unsubscribe")
	body.Add("hub.topic", topic)
	body.Add("hub.callback", callbackUrl)
	body.Add("hub.verify", "async")

	req, err := http.NewRequest("POST", hub, bytes.NewBufferString(body.Encode()))
	if err != nil {
		// error時は以降の処理をしない
		log.Infof(ctx, "SubscribeRequest err %v", err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	defer resp.Body.Close()
	reason, _ := ioutil.ReadAll(resp.Body)
	log.Infof(ctx, "reason %v", string(reason))
}
