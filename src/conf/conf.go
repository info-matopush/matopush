package conf

import (
	"golang.org/x/net/context"
	"github.com/mjibson/goon"
	"google.golang.org/appengine/datastore"
	"time"
        "src/site"
)

type SiteSubscribe struct {
	Id         string    `datastore:"-" goon:"id"`
	Endpoint   string    `datastore:"endpoint"`
	SiteUrl    string    `datastore:"site_url"`
	Value      string    `datastore:"value"`
	UpdateDate time.Time `datastore:"update_date,noindex"`
}

func Update(ctx context.Context, endpoint, siteUrl, value string) (string, error) {
	g := goon.FromContext(ctx)

	// todo: キーとなる文字列の長さが長すぎるかも。（優先度：低）
	ss := SiteSubscribe{
		Id: endpoint + ";" + siteUrl,
		Endpoint:endpoint,
		SiteUrl:siteUrl,
		Value:value,
		UpdateDate:time.Now(),
	}
	_, err := g.Put(&ss)
	if err != nil {
		return "", err
	}
	// サイトURLからサイト名への変換
	sui := site.SiteUpdateInfo{SiteUrl:ss.SiteUrl}
	err = g.Get(&sui)
	return sui.SiteTitle, err
}

func Cleanup(ctx context.Context, endpoint string) (error) {
	g := goon.FromContext(ctx)
	query := datastore.NewQuery("SiteSubscribe").Filter("endpoint=", endpoint)

	keys, err := g.GetAll(query, nil)
	if err == datastore.ErrInvalidEntityType {
		// エンティティがない場合
		return nil
	} else if err == nil {
		err = g.DeleteMulti(keys)
	}
	return err
}

