package webpush

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/info-matopush/matopush/conf"
	"github.com/info-matopush/matopush/endpoint"
	"github.com/info-matopush/matopush/site"
	"github.com/info-matopush/matopush/trace"
	"github.com/info-matopush/matopush/utility"

	"github.com/SherClockHolmes/webpush-go"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type pushMessage struct {
	FeedURL      string `json:"FeedUrl"`
	SiteURL      string `json:"SiteUrl"`
	SiteTitle    string
	SiteIcon     string
	ContentURL   string `json:"ContentUrl"`
	ContentTitle string
	ContentImage string
	Icon         string
	Endpoint     string
}

// TestHandler はEndpointに対してテスト通知を行う
func TestHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sui := site.UpdateInfo{
		Site: site.Site{
			SiteTitle: "まとプ　少し長めのタイトルの表示はこのようになります。",
			LatestContent: site.Content{
				Title: "これはテスト通知です。長めの文章の表示はこのようになります。ブラウザによっては画像も表示されます。",
				Image: "/img/IMGL5336_TP_V4.jpg",
			},
		},
	}

	ei, _ := endpoint.NewFromDatastore(ctx, r.FormValue("endpoint"))
	if ei != nil {
		sendPush(ctx, &sui, *ei)
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

// SendNotificationHandler はサイトの更新をPushで通知する
func SendNotificationHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params := r.URL.Query()
	feedURL := params.Get("FeedURL")
	if feedURL == "" {
		log.Errorf(ctx, "FeedURL is empty")
		return
	}
	log.Infof(ctx, "FeedURL:%v", feedURL)

	ui, _, err := site.FromURL(ctx, feedURL)
	if err != nil {
		log.Errorf(ctx, "FromURL error %v", err)
		return
	}

	err = ui.CheckSite(ctx)
	if err != nil {
		log.Errorf(ctx, "CheckSite error %v", err)
		return
	}

	// 更新があればPushを行う
	sendPushWhenSiteUpdate(ctx, ui)

	// 更新された情報を保存する
	ui.Update(ctx)
}

func sendPush(ctx context.Context, sui *site.UpdateInfo, ei endpoint.Endpoint) (err error) {
	// payloadの固定値はここで設定する
	m := pushMessage{
		FeedURL:      sui.FeedURL,
		SiteURL:      sui.SiteURL,
		SiteTitle:    sui.SiteTitle,
		SiteIcon:     sui.SiteIcon.TunneledURL(),
		ContentURL:   sui.LatestContent.URL,
		ContentTitle: sui.LatestContent.Title,
		ContentImage: sui.LatestContent.Image.TunneledURL(),
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

	pri, err := utility.GetPrivateKey(ctx)
	if err != nil {
		log.Errorf(ctx, "private key get error.%v", err)
		return
	}

	resp, err := webpush.SendNotification(message, &sub, &webpush.Options{
		HTTPClient:      client,
		Subscriber:      "https://matopush.appspot.com",
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
	confs := conf.GetAllFromFeedURL(ctx, sui.FeedURL)

	// 更新があれば通知
	if sui.UpdateFlg {
		log.Infof(ctx, "更新あり")

		var wg sync.WaitGroup
		for _, c := range confs {
			wg.Add(1)
			go func(ss conf.SiteSubscribe) {
				defer wg.Done()
				err = sendPush(ctx, sui, ss.Endpoint)
				if err == nil {
					trace.LogPush(ctx, ss.Endpoint.Endpoint, sui.SiteURL, sui.LatestContent.URL)
				}
			}(c)
		}
		wg.Wait()
	} else {
		log.Infof(ctx, "更新なし")
	}
	// 購読数を記録
	sui.Count = int64(len(confs))
	log.Infof(ctx, "url %v, count %v", sui.FeedURL, sui.Count)
	return
}

// HealthHandler は非表示のPushを全てのEndpointに送信し、
// 無効なEndpointを検出する
func HealthHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sui := site.UpdateInfo{
		Site: site.Site{
			SiteTitle: "まとプ",
			LatestContent: site.Content{
				Title: "",
			},
		},
	}

	sendPushAll(ctx, &sui)

	// 登録されているendpointの数を求める
	log.Infof(ctx, "有効なendpoint数. %d", endpoint.Count(ctx))
}

func sendPushAll(ctx context.Context, sui *site.UpdateInfo) {
	// 通知先のリストを取得する
	endpoints := endpoint.GetAll(ctx)

	if len(endpoints) > 0 {
		var wg sync.WaitGroup
		for _, e := range endpoints {
			wg.Add(1)
			go func(e endpoint.Endpoint) {
				defer wg.Done()
				sendPush(ctx, sui, e)
			}(e)
		}
		wg.Wait()
	}
	log.Debugf(ctx, "通知した数 %d", len(endpoints))
	return
}
