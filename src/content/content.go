package content

import (
	"time"

	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

const (
	maxContents int = 5
)

type pysicalContent struct {
	// KeyはコンテンツURLとする
	Key        string    `datastore:"-" goon:"id"`
	Title      string    `datastore:"title,noindex"`
	Summary    string    `datastore:"desc,noindex"`
	ImageURL   string    `datastore:"image_url,noindex"`
	ModifyDate time.Time `datastore:"modify_date,noindex"`
	CreateDate time.Time `datastore:"create_date,noindex"`
}

type Content struct {
	URL        string `json:"Url"`
	Title      string
	Summary    string
	ImageURL   string `json:"ImageUrl"`
	ModifyDate time.Time
}

// ContentFromFeed はフィードに含まれる
// コンテンツ(HTML)の情報
type ContentFromFeed struct {
	URL        string
	Title      string
	Summary    string
	ModifyDate time.Time
}

func New(ctx context.Context, cff ContentFromFeed) (*Content, error) {
	g := goon.FromContext(ctx)
	p := pysicalContent{Key: cff.URL}
	err := g.Get(&p)
	if err == datastore.ErrNoSuchEntity {
		return create(ctx, cff)
	}
	if err != nil {
		log.Infof(ctx, "goon get error %v, %v", cff.URL, err)
		return nil, err
	}
	c := p.makeContent()
	return &c, nil
}

func create(ctx context.Context, cff ContentFromFeed) (*Content, error) {
	h, err := HTMLParse(ctx, cff.URL)
	if err != nil {
		return nil, err
	}

	p := pysicalContent{
		Key:        cff.URL,
		Title:      cff.Title,
		Summary:    cff.Summary,
		ImageURL:   h.ImageURL,
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
		URL:        p.Key,
		Title:      p.Title,
		Summary:    p.Summary,
		ImageURL:   p.ImageURL,
		ModifyDate: p.ModifyDate,
	}
}

func Convert(ctx context.Context, cffs []ContentFromFeed) []Content {
	clist := []Content{}
	for _, cff := range cffs {
		c, err := New(ctx, cff)
		if err != nil {
			continue
		}
		clist = append(clist, *c)

		if len(clist) > maxContents {
			break
		}
	}
	return clist
}
