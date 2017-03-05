package main

import (
	"github.com/mjibson/goon"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"net/http"
	"src/endpoint"
	"src/site"
)

func menteHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	g := goon.NewGoon(r)

	// EndpointInfoをphysicalEndpointInfoにデータ移行
	var list []endpoint.EndpointInfo
	query := datastore.NewQuery("EndpointInfo")
	_, err := g.GetAll(query, &list)
	if err != nil {
		log.Errorf(ctx, "error. %v", err)
	}

	log.Infof(ctx, "mente num. %d", len(list))
	for _, src := range list {
		endpoint.Touch(appengine.NewContext(r), &src)
	}

	var siteList []site.SiteUpdateInfo
	query = datastore.NewQuery("SiteUpdateInfo")
	g.GetAll(query, &siteList)
	for _, s := range siteList {
		s.DeleteFlag = false
		g.Put(&s)
	}
}
