package conf

import (
	"encoding/base64"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"hash/fnv"
	"time"
)

// サイト購読情報
type SiteSubscribe struct {
	Endpoint string
	FeedUrl  string
	Enabled  bool
}

type physicalSiteSubscribe struct {
	Key        string    `datastore:"-" goon:"id"`
	Endpoint   string    `datastore:"endpoint"`
	FeedUrl    string    `datastore:"feed_url"`
	Enabled    bool      `datastore:"enabled"`
	UpdateDate time.Time `datastore:"update_date,noindex"`
	DeleteFlag bool      `datastore:"delete_flag"`
	DeleteDate time.Time `datastore:"delete_date,noindex"`
}

func makeKeyString(endpoint, feedUrl string) string {
	h := fnv.New64a()
	h.Write([]byte(endpoint + ";" + feedUrl))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// ユーザ固有設定(サイト購読情報)を更新する
func Update(ctx context.Context, endpoint, feedUrl string, enabled bool) error {
	g := goon.FromContext(ctx)
	ss := physicalSiteSubscribe{
		Key:        makeKeyString(endpoint, feedUrl),
		Endpoint:   endpoint,
		FeedUrl:    feedUrl,
		Enabled:    enabled,
		UpdateDate: time.Now(),
	}
	_, err := g.Put(&ss)
	return err
}

// 削除されたendpointの購読情報を物理削除する
func Cleanup(ctx context.Context, endpoint string) error {
	g := goon.FromContext(ctx)
	query := datastore.NewQuery("physicalSiteSubscribe").Filter("endpoint=", endpoint).KeysOnly()

	keys, err := g.GetAll(query, nil)
	log.Infof(ctx, "取得したSiteSubscribeの数 %d", len(keys))
	if err == datastore.ErrInvalidEntityType {
		// エンティティがない場合
		return nil
	}
	return g.DeleteMulti(keys)
}

// 購読情報の削除フラグを立てる
func (s *SiteSubscribe) Delete(ctx context.Context) {
	g := goon.FromContext(ctx)
	pss := physicalSiteSubscribe{
		Key: makeKeyString(s.Endpoint, s.FeedUrl),
	}
	err := g.Get(&pss)
	if err == nil {
		pss.DeleteFlag = true
		pss.DeleteDate = time.Now()
		pss.UpdateDate = time.Now()
		g.Put(&s)
	}
}

func ListFromEndpoint(ctx context.Context, endpoint string) []SiteSubscribe {
	g := goon.FromContext(ctx)

	query := datastore.NewQuery("physicalSiteSubscribe").Filter("endpoint=", endpoint)
	var confs []physicalSiteSubscribe
	var subs []SiteSubscribe
	_, err := g.GetAll(query, &confs)
	if err == nil {
		for _, conf := range confs {
			subs = append(subs, SiteSubscribe{
				Endpoint: conf.Endpoint,
				FeedUrl:  conf.FeedUrl,
				Enabled:  conf.Enabled,
			})
		}
	}
	return subs
}

func ListForPush(ctx context.Context, feedUrl string) []SiteSubscribe {
	g := goon.FromContext(ctx)

	query := datastore.NewQuery("physicalSiteSubscribe").Filter("feed_url=", feedUrl).Filter("enabled=", true)

	var confs []physicalSiteSubscribe
	var subs []SiteSubscribe
	_, err := g.GetAll(query, &confs)
	if err != nil {
		return subs
	}
	for _, conf := range confs {
		subs = append(subs, SiteSubscribe{
			Endpoint: conf.Endpoint,
			FeedUrl:  conf.FeedUrl,
			Enabled:  conf.Enabled,
		})
	}
	log.Infof(ctx, "conf.ListForPush %v, count %v", feedUrl, len(subs))
	return subs
}
