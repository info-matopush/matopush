package site

import (
	"encoding/json"
	"net/http"

	"google.golang.org/appengine"
)

// ListHandler は登録済（公開用）のサイトを返却する
func ListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	list := PublicList(ctx)
	b, _ := json.Marshal(list)
	w.Write(b)
}
