package content

import (
	"time"

	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

type pysicalContent struct {
	// KeyはコンテンツURLとする
	Key        string    `datastore:"-" goon:"id"`
	Title      string    `datastore:"title,noindex"`
	Summary    string    `datastore:"desc,noindex"`
	ImageUrl   string    `datastore:"image_url,noindex"`
	ModifyDate time.Time `datastore:"modify_date,noindex"`
	CreateDate time.Time `datastore:"create_date,noindex"`
}

type ContentFromFeed struct {
	Url        string
	Title      string
	Summary    string
	ModifyDate time.Time
}

type Content struct {
	Url        string
	Title      string
	Summary    string
	ImageUrl   string
	ModifyDate time.Time
}

type ContentInterface interface {
	GetContents() []Content
}

func New(ctx context.Context, cff ContentFromFeed) (*Content, error) {
	g := goon.FromContext(ctx)
	p := pysicalContent{Key: cff.Url}
	err := g.Get(&p)
	if err == datastore.ErrNoSuchEntity {
		return create(ctx, cff)
	}
	if err != nil {
		log.Infof(ctx, "goon get error %v, %v", cff.Url, err)
		return nil, err
	}
	c := p.makeContent()
	return &c, nil
}

func create(ctx context.Context, cff ContentFromFeed) (*Content, error) {
	h, err := htmlParse(ctx, cff.Url)
	if err != nil {
		return nil, err
	}

	p := pysicalContent{
		Key:        cff.Url,
		Title:      cff.Title,
		Summary:    cff.Summary,
		ImageUrl:   h.ImageUrl,
		ModifyDate: cff.ModifyDate,
		CreateDate: time.Now(),
	}
	g := goon.FromContext(ctx)
	g.Put(&p)
	c := p.makeContent()
	return &c, nil
}

func (p *pysicalContent) makeContent() Content {
	return Content{
		Url:        p.Key,
		Title:      p.Title,
		Summary:    p.Summary,
		ImageUrl:   p.ImageUrl,
		ModifyDate: p.ModifyDate,
	}
}
