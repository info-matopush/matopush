package src

import (
	"encoding/json"
	"net/http"

	"github.com/info-matopush/matopush/src/site"
	"google.golang.org/appengine"
)

// ListHandler は登録済（公開用）のサイトを返却する
func ListHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	list := site.PublicList(ctx)
	b, _ := json.Marshal(list)
	w.Write(b)
}
