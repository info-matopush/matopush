package src

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"

	"github.com/mjibson/goon"
	"golang.org/x/net/context"
)

// ServerKey はキーペアを保持する.
type ServerKey struct {
	Name  string `datastore:"-" goon:"id"`
	Value []byte `datastore:"value,noindex"`
}

// GetPublicKey は公開鍵を取得する.
func GetPublicKey(ctx context.Context) (*ecdsa.PublicKey, error) {
	pri, err := getPrivateKey(ctx)
	if err != nil {
		return nil, err
	}
	return &pri.PublicKey, nil
}

func getPrivateKey(ctx context.Context) (*ecdsa.PrivateKey, error) {
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
