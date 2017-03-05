package main

import (
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/http"
	"src/conf"
	"src/endpoint"
	"src/site"
)

func testHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sui := site.SiteUpdateInfo{SiteTitle: "まとプ", ContentTitle: "これはテスト通知です。"}

	ei, _ := endpoint.Get(ctx, r.FormValue("endpoint"))
	if ei != nil {
		sendPush(ctx, &sui, ei)
	}
}

func registHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	auth, _ := base64.RawURLEncoding.DecodeString(r.FormValue("auth"))
	p256dh, _ := base64.RawURLEncoding.DecodeString(r.FormValue("p256dh"))

	ei := &endpoint.EndpointInfo{
		Endpoint: r.FormValue("endpoint"),
		Auth:     auth,
		P256dh:   p256dh,
	}

	endpoint.Touch(ctx, ei)
}

func unregistHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	endpoint.Delete(ctx, r.FormValue("endpoint"))
}

func keyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	publicKey, err := GetPublicKey(ctx)
	if err != nil {
		log.Errorf(ctx, "get PublicKey error. %v", err)
		return
	}
	byteArray := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	fmt.Fprint(w, base64.RawURLEncoding.EncodeToString(byteArray))
}

func sendPush(ctx context.Context, sui *site.SiteUpdateInfo, ei *endpoint.EndpointInfo) (err error) {
	// payloadの固定値はここで設定する
	sui.Icon = "/img/news.png"
	sui.Endpoint = ei.Endpoint // ログ用

	message, _ := json.Marshal(sui)

	client := urlfetch.Client(ctx)
	b64 := base64.RawURLEncoding

	var sub webpush.Subscription
	sub.Endpoint = ei.Endpoint
	sub.Keys.Auth = b64.EncodeToString(ei.Auth)
	sub.Keys.P256dh = b64.EncodeToString(ei.P256dh)

	pri, err := getPrivateKey(ctx)
	if err != nil {
		log.Errorf(ctx, "private key get error.%v", err)
		return
	}

	resp, err := webpush.SendNotification(message, &sub, &webpush.Options{
		HTTPClient:      client,
		Subscriber:      "https://push2ch.appspot.com",
		TTL:             60,
		VAPIDPrivateKey: b64.EncodeToString(pri.D.Bytes()),
	})

	if err != nil {
		log.Errorf(ctx, "SendNotification error. %v", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode == 0 {
		log.Errorf(ctx, "send notification return code 0.")
		err = errors.New("send notification return code 0.")
	} else if resp.StatusCode == http.StatusOK {
		return
	} else if resp.StatusCode == http.StatusCreated {
		return
	} else if resp.StatusCode == http.StatusGone {
		endpoint.Delete(ctx, ei.Endpoint)
		err = errors.New("endpoint was gone.")
	} else {
		log.Infof(ctx, "resp %s", resp)
		buf, _ := ioutil.ReadAll(resp.Body)
		log.Infof(ctx, "body %v", string(buf))
		err = errors.New("unknown response.")
	}
	return
}

func sendPushWhenSiteUpdate(ctx context.Context, sui *site.SiteUpdateInfo) (err error) {
	err = nil

	// 更新がなければ通知はしない
	if !sui.UpdateFlg {
		return
	}

	// 通知先のリストを取得する
	g := goon.FromContext(ctx)
	query := datastore.NewQuery("SiteSubscribe").Filter("site_url=", sui.SiteUrl).Filter("value=", "true").Limit(1000)
	it := g.Run(query)
	// 購読数カウントクリア
	sui.SubscribeCount = 0
	for {
		var ss conf.SiteSubscribe
		_, err := it.Next(&ss)
		if err == datastore.Done {
			// エラーとして扱わない
			err = nil
			break
		}
		if err != nil {
			log.Errorf(ctx, "datastore get error.%v", err)
			break
		}

		ei, err := endpoint.Get(ctx, ss.Endpoint)
		if err != nil {
			// endpointが見つからなかった場合(cleanupミス？)はSiteSubscribeの削除フラグを立てる
			conf.Delete(ctx, ss)
			continue
		}

		sui.SubscribeCount++
		sendPush(ctx, sui, ei)
	}
	// 購読数を記録
	g.Put(sui)
	return
}

func sendPushAll(ctx context.Context, sui *site.SiteUpdateInfo) {
	// 通知先のリストを取得する
	list := endpoint.GetAll(ctx)

	for _, ei := range list {
		sendPush(ctx, sui, &ei)
	}
	log.Debugf(ctx, "通知した数 %d", len(list))
	return
}
