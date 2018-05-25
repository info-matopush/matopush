package src

import (
	"io"
	"net/http"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

// TunnelHandler はhttpで提供されているコンテンツにアクセスするためのproxy機能
func TunnelHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	params := r.URL.Query()
	url := params.Get("url")
	if url == "" {
		log.Errorf(ctx, "url is empty")
		w.WriteHeader(http.StatusForbidden)
		return
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// error時は以降の処理をしない
		log.Infof(ctx, "TunnelHandler:http.NewRequest err %v, url %v", err, url)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	client := urlfetch.Client(ctx)
	resp, err := client.Do(req)
	if err != nil {
		log.Infof(ctx, "TunnelHandler:client.Do err %v, url %v", err, url)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	defer resp.Body.Close()
	copyHeaders(w.Header(), resp.Header)
	io.Copy(w, resp.Body)
	w.WriteHeader(resp.StatusCode)
}

func copyHeaders(dst, src http.Header) {
	for k := range dst {
		dst.Del(k)
	}
	for k, vs := range src {
		for _, v := range vs {
			dst.Add(k, v)
		}
	}
}
