package content

import (
	"time"

	"github.com/info-matopush/matopush/src/remodel"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const (
	maxContents int = 5
)

type physicalContent struct {
	// KeyはコンテンツURLとする
	Key        string    `datastore:"-" goon:"id"`
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

// NewFromFeed はフィードで得た情報を元にコンテンツ情報を作成する
func NewFromFeed(ctx context.Context, ff FromFeed) (*Content, error) {
	g := goon.FromContext(ctx)
	p := physicalContent{Key: ff.URL}
	err := g.Get(&p)
	if err == datastore.ErrNoSuchEntity {
		return create(ctx, ff)
	}
	if err != nil {
		log.Infof(ctx, "goon get error %v, %v", ff.URL, err)
		return nil, err
	}
	c := p.makeContent()
	return &c, nil
}

func create(ctx context.Context, ff FromFeed) (*Content, error) {
	h, err := ParseHTML(ctx, ff.URL)
	if err != nil {
		return nil, err
	}

	p := physicalContent{
		Key:        ff.URL,
		Title:      ff.Title,
		Summary:    ff.Summary,
		ImageURL:   h.ImageURL,
		ModifyDate: ff.ModifyDate,
		CreateDate: time.Now(),
	}
	g := goon.FromContext(ctx)
	g.Put(&p)
	c := p.makeContent()
	return &c, nil
}

func (p *physicalContent) makeContent() Content {
	return Content{
		URL:        p.Key,
		Title:      p.Title,
		Summary:    p.Summary,
		ImageURL:   remodel.ExURL(p.ImageURL),
		ModifyDate: p.ModifyDate,
	}
}

// Convert はFromFeed配列をContent配列に変換する
func Convert(ctx context.Context, ffs []FromFeed) []Content {
	clist := []Content{}
	for _, ff := range ffs {
		c, err := NewFromFeed(ctx, ff)
		if err != nil {
			log.Warningf(ctx, "contents New error %v", err)
			continue
		}
		clist = append(clist, *c)

		if len(clist) > maxContents {
			break
		}
	}
	return clist
}
