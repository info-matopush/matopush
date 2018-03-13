package main

import (
	"google.golang.org/appengine"
	"net/http"
	"src/conf"
)

func mainteHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	conf.Migration(ctx)
}
