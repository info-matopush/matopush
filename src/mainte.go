package src

import (
	"net/http"

	"github.com/info-matopush/matopush/src/conf"
	"google.golang.org/appengine"
)

// MainteHandler はメンテナンス用の処理を行う
func MainteHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	subs := conf.GetAll(ctx)

	for _, sub := range subs {
		err := conf.Update(ctx, sub.Endpoint.Endpoint, sub.FeedURL, sub.Enabled)
		if err != nil {
			conf.Delete(ctx, sub.Endpoint.Endpoint, sub.FeedURL)
		}
	}
}
