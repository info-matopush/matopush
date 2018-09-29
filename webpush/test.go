package webpush

import (
	"net/http"

	"github.com/info-matopush/matopush/endpoint"
	"github.com/info-matopush/matopush/site"
	"github.com/info-matopush/matopush/utility"
	"google.golang.org/appengine"
)

var testPush = site.ForPush{
	SiteTitle:    "まとプ　少し長めのタイトルの表示はこのようになります。",
	ContentTitle: "これはテスト通知です。長めの文章の表示はこのようになります。ブラウザによっては画像も表示されます。",
	ContentImage: utility.ExURL("/img/IMGL5336_TP_V4.jpg"),
}

// TestHandler はEndpointに対してテスト通知を行う
func TestHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	ei, _ := endpoint.NewFromDatastore(ctx, r.FormValue("endpoint"))
	if ei != nil {
		sendPush(ctx, testPush, *ei)
	}
}
