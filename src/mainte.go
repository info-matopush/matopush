package src

import (
	"net/http"

	"github.com/info-matopush/matopush/src/site"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

func MainteHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	siteList, err := site.List(ctx)
	if err != nil {
		log.Errorf(ctx, "get site failed. %v", err)
	}

	for _, ui := range siteList {
		if ui.HubUrl != "" {
			SubscribeRequest(ctx, SubscribeURL+ui.FeedUrl, ui.FeedUrl, ui.HubUrl, ui.Secret)
		}
	}
}
