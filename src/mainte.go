package src

import (
	"net/http"

	"github.com/info-matopush/matopush/src/content"
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
		if ui.SiteIcon == "" {
			h, err := content.HTMLParse(ctx, ui.SiteURL)
			if err != nil {
				log.Infof(ctx, "HTMLParse error %v", err)
				continue
			}
			ui.SiteIcon = h.IconURL
			ui.Update(ctx)
		}
	}
}

func SubscribeRequestHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	siteList, err := site.List(ctx)
	if err != nil {
		log.Errorf(ctx, "get site failed. %v", err)
	}

	for _, ui := range siteList {
		if ui.HubURL != "" {
			SubscribeRequest(ctx, SubscribeURL+ui.FeedURL, ui.FeedURL, ui.HubURL, ui.Secret)
		}
	}
}
