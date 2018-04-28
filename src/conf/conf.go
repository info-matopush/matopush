package conf

import (
	"encoding/base64"
	"hash/fnv"
	"time"

	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// SiteSubscribe はサイト購読情報
type SiteSubscribe struct {
	Endpoint string
	FeedURL  string
	Enabled  bool
}

type physicalSiteSubscribe struct {
	Key        string    `datastore:"-" goon:"id"`
	Endpoint   string    `datastore:"endpoint"`
	FeedURL    string    `datastore:"feed_url"`
	Enabled    bool      `datastore:"enabled"`
	UpdateDate time.Time `datastore:"update_date,noindex"`
	DeleteFlag bool      `datastore:"delete_flag"`
	DeleteDate time.Time `datastore:"delete_date,noindex"`
}

func makeKeyString(endpoint, FeedURL string) string {
	h := fnv.New64a()
	h.Write([]byte(endpoint + ";" + FeedURL))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Delete は購読情報を削除する
func Delete(ctx context.Context, endpoint, feedURL string) {
	g := goon.FromContext(ctx)
	ss := physicalSiteSubscribe{
		Key: makeKeyString(endpoint, feedURL),
	}
	err := g.Delete(g.Key(&ss))
	if err != nil {
		log.Errorf(ctx, "g.Delete error: %v", err)
	}
}

// Update はユーザ固有設定(サイト購読情報)を更新する
func Update(ctx context.Context, endpoint, feedURL string, enabled bool) error {
	g := goon.FromContext(ctx)
	ss := physicalSiteSubscribe{
		Key:        makeKeyString(endpoint, feedURL),
		Endpoint:   endpoint,
		FeedURL:    feedURL,
		Enabled:    enabled,
		UpdateDate: time.Now(),
	}
	_, err := g.Put(&ss)
	return err
}

// Cleanup は削除されたendpointの購読情報を物理削除する
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

// Delete は購読情報の削除フラグを立てる
func (s *SiteSubscribe) Delete(ctx context.Context) {
	g := goon.FromContext(ctx)
	pss := physicalSiteSubscribe{
		Key: makeKeyString(s.Endpoint, s.FeedURL),
	}
	err := g.Get(&pss)
	if err == nil {
		pss.DeleteFlag = true
		pss.DeleteDate = time.Now()
		pss.UpdateDate = time.Now()
		g.Put(&s)
	}
}

// ListFromEndpoint はendpointに紐づくサイト購読情報を取得する
func ListFromEndpoint(ctx context.Context, endpoint string) []SiteSubscribe {
	g := goon.FromContext(ctx)

	query := datastore.NewQuery("physicalSiteSubscribe").Filter("endpoint=", endpoint).Filter("delete_flag=", false)
	var confs []physicalSiteSubscribe
	var subs []SiteSubscribe
	_, err := g.GetAll(query, &confs)
	if err == nil {
		for _, conf := range confs {
			subs = append(subs, SiteSubscribe{
				Endpoint: conf.Endpoint,
				FeedURL:  conf.FeedURL,
				Enabled:  conf.Enabled,
			})
		}
	}
	return subs
}

// ListFromFeedURL はfeedURLに紐づく有効なサイト購読情報を全て取得する
func ListFromFeedURL(ctx context.Context, feedURL string) []SiteSubscribe {
	g := goon.FromContext(ctx)

	query := datastore.NewQuery("physicalSiteSubscribe").Filter("feed_url=", feedURL).Filter("enabled=", true)

	var confs []physicalSiteSubscribe
	var subs []SiteSubscribe
	_, err := g.GetAll(query, &confs)
	if err != nil {
		return subs
	}
	for _, conf := range confs {
		subs = append(subs, SiteSubscribe{
			Endpoint: conf.Endpoint,
			FeedURL:  conf.FeedURL,
			Enabled:  conf.Enabled,
		})
	}
	log.Infof(ctx, "conf.ListForPush %v, count %v", feedURL, len(subs))
	return subs
}
