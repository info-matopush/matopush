package utility

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

// ExURL は拡張URL型
type ExURL string

// MarshalJSON は拡張ExURLをJSON出力する
func (e ExURL) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.TunneledURL())
}

// TunneledURL はproxyされたURLを返却する
func (e ExURL) TunneledURL() string {
	if strings.HasPrefix(string(e), "http://") {
		return "/api/tunnel?url=" + string(e)
	}
	return string(e)
}

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
	// キャッシュを有効にする
	w.Header().Set("cache-control", "public, max-age=259200")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
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
