package main

import (
	"net/http"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"encoding/xml"
	"io/ioutil"
	"src/atom"
)

func dummyHander(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)


	client := urlfetch.Client(ctx)
	resp, err := client.Get("http://blog.livedoor.jp/corez18c24-mili777")
	if err != nil {
		log.Infof(ctx, "site get error. %v", err)
		return
	}

	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	log.Infof(ctx, "body. %s", string(body))

	result := atom.Feed{}
	err = xml.Unmarshal(body, &result)

	log.Infof(ctx, "rss %v, err %v", result, err)
	if len(result.Entry) > 0 {
		log.Infof(ctx, "entry[0]. %v", result.Entry[0])
	}
}
