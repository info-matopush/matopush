package webpush

import (
	"net/http"
	"sync"

	"github.com/info-matopush/matopush/endpoint"
	"github.com/info-matopush/matopush/site"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

var healthPush = site.ForPush{
	SiteTitle:    "まとプ",
	ContentTitle: "",
}

// HealthHandler は非表示のPushを全てのEndpointに送信し、
// 無効なEndpointを検出する
func HealthHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	sendPushAll(ctx, healthPush)
}

func sendPushAll(ctx context.Context, site site.ForPush) {
	// 通知先のリストを取得する
	endpoints := endpoint.GetAll(ctx)

	if len(endpoints) > 0 {
		var wg sync.WaitGroup
		for _, e := range endpoints {
			wg.Add(1)
			go func(e endpoint.Endpoint) {
				defer wg.Done()
				sendPush(ctx, site, e)
			}(e)
		}
		wg.Wait()
	}
	log.Debugf(ctx, "通知した数 %d", len(endpoints))
	return
}
