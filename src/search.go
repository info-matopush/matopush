package main

import (
	"encoding/json"
	"golang.org/x/oauth2/google"
	customsearch "google.golang.org/api/customsearch/v1"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"net/http"
	"strconv"
)

var j = `{
  "type": "service_account",
  "project_id": "matopush",
  "private_key_id": "d72e2e30905c8186be593af2d5a464ca7291b635",
  "private_key": "-----BEGIN PRIVATE KEY-----\nMIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCytolXe2tiWQeb\nGlNpazNBOQMpOG+mjbtnx360/TQi3wWjLSxjHlvgXpjSh4fOaDk23hCoJ9em0ZPy\nuXL87/dp0njfq3vd4cHO3Yjdy71m/8WecfHPR0ykcc18D/QolKSXSRLyxKR01dtF\nrSlcJa3qjkEbcUwVMyBWNixXPYhYYxftu/Mm4WohhiQm2pytzDVKtsx7zEZ4XAUh\nUSzyJhFjlhjTkJoB2qkI/Geiz5jN7e8XTtCvTyK0ODixr1fCJmh19TyBHLDHj13q\naJUXHEoC2lxn4Q8ITW5FQV/Qp87Ssaq8RS+u4UhSFtqrASu1XPFzRc6c8TiW2F0T\nJ75to6KJAgMBAAECggEACLhX0T+OomwRSL2RgY/sHoWapHtCSQOanlpUnp4wQHdW\nl65PmSW7elVptydLA1GAlGKXhRNhaWljT7FCB03qNtN0ETqvC5X4oyoAg9oG8hMO\n4Hrs2QSNaWshvKj50xt8DUm3TyzOXXxF4iUSr8Vcf0wIdaW8mAtrmE+1lpHEQmCJ\nvolJX+VyPUQIPS7TSpiK3+tkU5SfzTXpni0jXm7jUW9wwLz0IGPtqQNBpa5bRRLS\nF7uxccf02PUmW+U+DLWiCf666vIByaKN8CZ97b/hs3ggHRt2n9JGMzAcRDiObaMY\nVq6WVOOaIHLgRU1+S2A+1+h4EsMbqRRzzrRodirk1QKBgQDf4HUNSuz9l4rGlPT6\n36P82ZgIdrI4gvm6SQYoHK54KgIU0vahUBbL0kP3HNDnxLNSqYgjjHea7vo+7en7\nei1w/rMgaWjZMmX44vdEUD9gJN5SbHAGltgB4lY7HLtnAnfURd6D5uxQx02tnWz2\nWI0hanlcv/p5AIhmsfdW6dAkvwKBgQDMWxqeWvwEstourCZf5vzLAG3roiRFXq3g\nx+a69hvsZGk0AUwjBBbccy9I0jT9IkuBX7g0Q1gcyrO6Lm+obu/L5cz67YDmhH+2\nTv5Rcin0V543Qjpgex8XH3B2ccvRtfEmU8ZyobZAJeJeJP1qwPMNItbq3zDd1ToB\nHEbGQQgitwKBgFqGZE5PsayJDnBl4vleXOzs/3DMrhvzug79YCPwFQw50EWjWF66\nB7269AiD+mT9QJV4P7hAIEzhvQadJTOun5lFJCFC/kZ0/o65F8rjt/yka9FgT5wa\nepWoc73LTGvGr7WB2wvy4DN5o4tEUL7753VPnGtIpXswH/eGlsDqImP5AoGBAKV8\nFSctSK5JY1OuRnkc5ZNCasEJEVQ3opjHaHn4OI6KlYLulgg5FIY6pIzU5OIj9n7y\n04lHC8BtCXP4jKUaCQfVtNNypxKFM6Kff2TXDVB374CSGhHtQjUIWZsg9cuCCaFe\n7/H+MEbsJs7UJ39edrQphV63lKvfMtSZYFrFaOArAoGBANd3rd8R2XsG9zN/LLra\nohtFKhNBuK0IzfqIvEijPWddA9vsGzvARx7mbnLYMidKUXSb9PDBXI1F1Z/tnPlS\nBKEpkynHTK9+/FvLXevk7S4v0dl/5+y8jXOkfRayCXzjUvTJvAyDtJnF3aS6ntvj\n9QA/a2smW70BTfeuuxKCnNdS\n-----END PRIVATE KEY-----\n",
  "client_email": "matopush@appspot.gserviceaccount.com",
  "client_id": "116765465910511464687",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://accounts.google.com/o/oauth2/token",
  "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
  "client_x509_cert_url": "https://www.googleapis.com/robot/v1/metadata/x509/matopush%40appspot.gserviceaccount.com"
}`

func searchHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	p := NewFromContext(ctx)
	//key := p.GetString("google.custom.search.apikey", "undefined")
	id := p.GetString("google.search.engine.id", "undefined")

	log.Infof(ctx, "google.search.engine.id: %v", id)

	// endpoint := r.FormValue("endpoint")
	keyword := r.FormValue("keyword")
	pos, err := strconv.Atoi(r.FormValue("position"))
	if err != nil {
		pos = 1
	}

	conf, err := google.JWTConfigFromJSON([]byte(j), "https://www.googleapis.com/auth/cse")
	if err != nil {
		log.Infof(ctx, "JWTConfigFromJSON error. %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}

	cseService, err := customsearch.New(conf.Client(ctx))
	if err != nil {
		log.Infof(ctx, "customsearch.New err. %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	search := cseService.Cse.List(keyword)
	search.Cx(id)
	search.Start(int64(pos))

	s, err := search.Do()
	if err != nil {
		log.Infof(ctx, "search.Do error. keyword %v, %v", keyword, err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	b, err := json.Marshal(s)
	if err != nil {
		log.Infof(ctx, "json.Marshal error. %v", err)
		w.WriteHeader(http.StatusForbidden)
		return
	}
	w.Write(b)
}
