package conf

import (
	"encoding/base64"
	"hash/fnv"
	"time"

	"github.com/info-matopush/matopush/src/endpoint"
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
	P256dh     []byte    `datastore:"p256dh,noindex"`
	Auth       []byte    `datastore:"auth,noindex"`
	FeedURL    string    `datastore:"feed_url"`
	Enabled    bool      `datastore:"enabled"`
	UpdateDate time.Time `datastore:"update_date,noindex"`
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
func Update(ctx context.Context, e, feedURL string, enabled bool) error {
	ep, err := endpoint.NewFromDatastore(ctx, e)
	if err != nil {
		return err
	}

	g := goon.FromContext(ctx)
	ss := physicalSiteSubscribe{
		Key:        makeKeyString(ep.Endpoint, feedURL),
		Endpoint:   ep.Endpoint,
		P256dh:     ep.P256dh,
		Auth:       ep.Auth,
		FeedURL:    feedURL,
		Enabled:    enabled,
		UpdateDate: time.Now(),
	}
	_, err = g.Put(&ss)
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

// Delete は購読情報を削除する
func (s *SiteSubscribe) Delete(ctx context.Context) {
	Delete(ctx, s.Endpoint, s.FeedURL)
}

// ListFromEndpoint はendpointに紐づくサイト購読情報を取得する
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
				FeedURL:  conf.FeedURL,
				Enabled:  conf.Enabled,
			})
		}
	}
	return subs
}

// EndpointsFromFeedURL はfeedURLに紐づく有効なEndpointを全て取得する
func EndpointsFromFeedURL(ctx context.Context, feedURL string) (dst []endpoint.Endpoint) {
	g := goon.FromContext(ctx)

	query := datastore.NewQuery("physicalSiteSubscribe").Filter("feed_url=", feedURL).Filter("enabled=", true)

	var confs []physicalSiteSubscribe
	var subs []SiteSubscribe
	_, err := g.GetAll(query, &confs)
	if err == nil {
		return
	}
	for _, conf := range confs {
		dst = append(dst, endpoint.Endpoint{
			Endpoint: conf.Endpoint,
			P256dh:   conf.P256dh,
			Auth:     conf.Auth,
		})
	}
	log.Infof(ctx, "conf.ListFromFeedURL %v, count %v", feedURL, len(subs))
	return
}

// GetAll はDatastore上のSiteSubscribeを全て返却する
func GetAll(ctx context.Context) (dst []SiteSubscribe) {
	g := goon.FromContext(ctx)

	query := datastore.NewQuery("physicalSiteSubscribe")

	var confs []physicalSiteSubscribe
	var subs []SiteSubscribe
	_, err := g.GetAll(query, &confs)
	if err != nil {
		return subs
	}
	for _, conf := range confs {
		dst = append(dst, SiteSubscribe{
			Endpoint: conf.Endpoint,
			FeedURL:  conf.FeedURL,
			Enabled:  conf.Enabled,
		})
	}
	log.Debugf(ctx, "conf.GetAll count %d", len(subs))
	return
}
