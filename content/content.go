package content

import (
	"encoding/base64"
	"hash/fnv"
	"sync"
	"time"

	"github.com/info-matopush/matopush/remodel"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const (
	maxContents int = 5
)

type physicalContent struct {
	// KeyはコンテンツURLをハッシュ化したものとする
	Key        string    `datastore:"-" goon:"id"`
	URL        string    `datastore:"url,noindex"`
	Title      string    `datastore:"title,noindex"`
	Summary    string    `datastore:"desc,noindex"`
	ImageURL   string    `datastore:"image_url,noindex"`
	ModifyDate time.Time `datastore:"modify_date,noindex"`
	CreateDate time.Time `datastore:"create_date,noindex"`
}

// Content コンテンツ情報
type Content struct {
	URL        string `json:"Url"`
	Title      string
	Summary    string
	ImageURL   remodel.ExURL `json:"ImageUrl"`
	ModifyDate time.Time
}

// FromFeed はフィードに含まれる
// コンテンツ(HTML)の情報
type FromFeed struct {
	URL        string
	Title      string
	Summary    string
	ModifyDate time.Time
}

// URLは長すぎる場合があるので、ハッシュを使ってキーを作成する
func strToKeyString(str string) string {
	h := fnv.New128a()
	h.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// newFromFeed はフィードで得た情報を元にコンテンツ情報を作成する
func newFromFeed(ctx context.Context, ff FromFeed) Content {
	g := goon.FromContext(ctx)
	p := physicalContent{
		Key: strToKeyString(ff.URL),
	}
	err := g.Get(&p)
	if err == datastore.ErrNoSuchEntity {
		// Datastore上にデータがない場合は、physicalContentを作成、保存を行う。
		p = create(ctx, ff)
	} else if err != nil {
		log.Warningf(ctx, "NewFromFeed g.Get error %v, %v", ff, err)
		p = create(ctx, ff)
	}
	return p.makeContent()
}

func create(ctx context.Context, ff FromFeed) physicalContent {
	p := physicalContent{
		Key:        strToKeyString(ff.URL),
		URL:        ff.URL,
		Title:      ff.Title,
		Summary:    ff.Summary,
		ImageURL:   "",
		ModifyDate: ff.ModifyDate,
		CreateDate: time.Now(),
	}
	h, err := ParseHTML(ctx, ff.URL)
	if err == nil {
		// HTMLの取得に成功したらImageURLを設定する
		p.ImageURL = h.ImageURL
	}

	g := goon.FromContext(ctx)
	_, err = g.Put(&p)
	if err != nil {
		log.Warningf(ctx, "create g.Put error %v, %v", ff, err)
	}
	return p
}

func (p *physicalContent) makeContent() Content {
	return Content{
		URL:        p.URL,
		Title:      p.Title,
		Summary:    p.Summary,
		ImageURL:   remodel.ExURL(p.ImageURL),
		ModifyDate: p.ModifyDate,
	}
}

// Convert はFromFeed配列をContent配列に変換する
func Convert(ctx context.Context, ffs []FromFeed) []Content {
	if len(ffs) > maxContents {
		ffs = ffs[:maxContents]
	}
	clist := make([]Content, len(ffs))
	var wg sync.WaitGroup
	for i, ff := range ffs {
		wg.Add(1)
		go func(i int, ff FromFeed) {
			defer wg.Done()
			c := newFromFeed(ctx, ff)
			clist[i] = c
		}(i, ff)
	}
	wg.Wait()
	return clist
}
