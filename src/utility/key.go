package utility

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
)

// KeyHandler は公開鍵を返却する
func KeyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := appengine.NewContext(r)

	publicKey, err := GetPublicKey(ctx)
	if err != nil {
		log.Errorf(ctx, "get public key error: %v", err)
		return
	}
	byteArray := elliptic.Marshal(publicKey.Curve, publicKey.X, publicKey.Y)
	fmt.Fprint(w, base64.RawURLEncoding.EncodeToString(byteArray))
}

// ServerKey はキーペアを保持する
type ServerKey struct {
	Name  string `datastore:"-" goon:"id"`
	Value []byte `datastore:"value,noindex"`
}

// GetPublicKey は公開鍵を取得する
func GetPublicKey(ctx context.Context) (*ecdsa.PublicKey, error) {
	pri, err := GetPrivateKey(ctx)
	if err != nil {
		return nil, err
	}
	return &pri.PublicKey, nil
}

// GetPrivateKey は秘密鍵を取得する
func GetPrivateKey(ctx context.Context) (*ecdsa.PrivateKey, error) {
	g := goon.FromContext(ctx)

	pri := ServerKey{Name: "PrivateKey"}
	g.Get(&pri)
	if pri.Value == nil {
		key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, err
		}
		pri.Value, err = x509.MarshalECPrivateKey(key)
		if err != nil {
			return nil, err
		}
		g.Put(&pri)
	}
	return x509.ParseECPrivateKey(pri.Value)
}
