package webpush

import (
	"encoding/base64"
	"net/http"

	"github.com/info-matopush/matopush/endpoint"
	"google.golang.org/appengine"
)

// RegistHandler はEndpointを登録する
func RegistHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	auth, _ := base64.RawURLEncoding.DecodeString(r.FormValue("auth"))
	p256dh, _ := base64.RawURLEncoding.DecodeString(r.FormValue("p256dh"))

	ei := &endpoint.Endpoint{
		Endpoint: r.FormValue("endpoint"),
		Auth:     auth,
		P256dh:   p256dh,
	}

	ei.Touch(ctx)
}

// UnregistHandler はEndpointを解除する
func UnregistHandler(_ http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)
	e, err := endpoint.NewFromDatastore(ctx, r.FormValue("endpoint"))
	if err != nil {
		return
	}
	e.Delete(ctx)
}
