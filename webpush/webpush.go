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

// SendNotificationHandler はサイトの更新をPushで通知する
func SendNotificationHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params := r.URL.Query()
	feedURL := params.Get("FeedURL")
	if feedURL == "" {
		log.Errorf(ctx, "FeedURL is empty")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// 通知先のリストを取得する
	endpoints := conf.GetAllFromFeedURL(ctx, feedURL)

	// 通知情報を取得する
	site, err := site.GetSiteForPush(ctx, feedURL)
	if err != nil {
		log.Errorf(ctx, "site.GetSiteForPush error: %v", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var wg sync.WaitGroup
	for _, endpoint := range endpoints {
		wg.Add(1)
		go func(endpoint conf.SiteSubscribe) {
			defer wg.Done()
			err := sendPush(ctx, site, endpoint.Endpoint)
			if err == nil {
				trace.LogPush(ctx, endpoint.Endpoint.Endpoint, site.SiteURL, site.ContentURL)
			}
		}(endpoint)
	}
	wg.Wait()
	log.Debugf(ctx, "通知した数 %d", len(endpoints))
}

func sendPush(ctx context.Context, site site.ForPush, ei endpoint.Endpoint) (err error) {
	// payloadの固定値はここで設定する
	m := pushMessage{
		FeedURL:      site.FeedURL,
		SiteURL:      site.SiteURL,
		SiteTitle:    site.SiteTitle,
		SiteIcon:     site.SiteIcon.TunneledURL(),
		ContentURL:   site.ContentURL,
		ContentTitle: site.ContentTitle,
		ContentImage: site.ContentImage.TunneledURL(),
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
