package src

import (
	"crypto/elliptic"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/info-matopush/matopush/src/endpoint"
	"github.com/info-matopush/matopush/src/site"

	"github.com/SherClockHolmes/webpush-go"
	"github.com/info-matopush/matopush/src/conf"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type pushMessage struct {
	FeedURL      string `json:"FeedUrl"`
	SiteURL      string `json:"SiteUrl"`
	SiteTitle    string
	ContentURL   string `json:"ContentUrl"`
	ContentTitle string
	Icon         string
	Endpoint     string
}

// TestHandler はEndpointに対してテスト通知を行う
func TestHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sui := site.UpdateInfo{SiteTitle: "まとプ", ContentTitle: "これはテスト通知です。"}

	ei, _ := endpoint.NewFromDatastore(ctx, r.FormValue("endpoint"))
	if ei != nil {
		sendPush(ctx, &sui, ei)
	}
}

// RegistHandler はEndpointを登録する
func RegistHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	auth, _ := base64.RawURLEncoding.DecodeString(r.FormValue("auth"))
	p256dh, _ := base64.RawURLEncoding.DecodeString(r.FormValue("p256dh"))

	ei := &endpoint.Endpoint{
		Endpoint: r.FormValue("endpoint"),
		Auth:     auth,
		P256dh:   p256dh,
	}

	ei.Touch(ctx)
}

// UnregistHandler はEndpointを解除する
func UnregistHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	e, err := endpoint.NewFromDatastore(ctx, r.FormValue("endpoint"))
	if err != nil {
		return
	}
	e.Delete(ctx)
}

// KeyHandler は公開鍵を返却する
func KeyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	publicKey, err := GetPublicKey(ctx)
	if err != nil {
		log.Errorf(ctx, "get public key error: %v", err)
		return
	}
	byteArray := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	fmt.Fprint(w, base64.RawURLEncoding.EncodeToString(byteArray))
}

func sendPush(ctx context.Context, sui *site.UpdateInfo, ei *endpoint.Endpoint) (err error) {
	// payloadの固定値はここで設定する
	m := pushMessage{
		FeedURL:      sui.FeedURL,
		SiteURL:      sui.SiteURL,
		SiteTitle:    sui.SiteTitle,
		ContentURL:   sui.ContentURL,
		ContentTitle: sui.ContentTitle,
		Icon:         "/img/news.png",
		Endpoint:     ei.Endpoint, // クライアントでログするのに使用
	}
	message, err := json.Marshal(m)
	if err != nil {
		return
	}

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
		err = errors.New("send notification return code 0")
	} else if resp.StatusCode == http.StatusOK {
		return
	} else if resp.StatusCode == http.StatusCreated {
		return
	} else if resp.StatusCode == http.StatusGone {
		ei.Delete(ctx)
		err = errors.New("endpoint was gone")
	} else {
		log.Infof(ctx, "resp %s", resp)
		buf, _ := ioutil.ReadAll(resp.Body)
		log.Infof(ctx, "body %v", string(buf))
		err = errors.New("unknown response")
	}
	return
}

func sendPushWhenSiteUpdate(ctx context.Context, sui *site.UpdateInfo) (err error) {
	// 通知先のリストを取得する
	s := conf.ListFromFeedURL(ctx, sui.FeedURL)

	// 更新があれば通知
	if sui.UpdateFlg {
		for _, ss := range s {
			ei, err := endpoint.NewFromDatastore(ctx, ss.Endpoint)
			if err != nil {
				// endpointが見つからなかった場合(cleanupミス？)はSiteSubscribeの削除フラグを立てる
				ss.Delete(ctx)
				continue
			}

			err = sendPush(ctx, sui, ei)
			if err == nil {
				// LogPush(ctx, sui.Endpoint, sui.SiteUrl, sui.ContentUrl)
			}
		}
	}
	// 購読数を記録
	sui.Count = int64(len(s))
	log.Infof(ctx, "url %v, count %v", sui.FeedURL, sui.Count)
	return
}

func sendPushAll(ctx context.Context, sui *site.UpdateInfo) {
	// 通知先のリストを取得する
	list := endpoint.GetAll(ctx)

	for _, ei := range list {
		sendPush(ctx, sui, &ei)
	}
	log.Debugf(ctx, "通知した数 %d", len(list))
	return
}
