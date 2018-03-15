package main

import (
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
	"src/site"
)

func mainteHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	siteList, err := site.List(ctx)
	if err != nil {
		log.Errorf(ctx, "get site failed. %v", err)
	}

	for _, ui := range siteList {
		if ui.HubUrl != "" {
			SubscribeRequest(ctx, SubscribeUrl+ui.FeedUrl, ui.FeedUrl, ui.HubUrl, ui.Secret)
		}
	}
}
