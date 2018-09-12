package site

import (
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/info-matopush/matopush/content"
	"github.com/info-matopush/matopush/utility"
	"github.com/info-matopush/matopush/xml/atom"
	"github.com/info-matopush/matopush/xml/rdf"
	"github.com/info-matopush/matopush/xml/rss"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

// KeyはFeedUrl
type physicalSite struct {
	Key           string            `datastore:"-" goon:"id"`
	Type          string            `datastore:"type,noindex"`
	SiteURL       string            `datastore:"site_url,noindex"`
	SiteTitle     string            `datastore:"site_title,noindex"`
	SiteIcon      string            `datastore:"site_icon,noindex"`
	LatestContent Content           `datastore:"latest,noindex"`
	Public        bool              `datastore:"public"`
	HubURL        string            `datastore:"hub_url,noindex"`
	Count         int64             `datastore:"count,noindex"`
	CreateDate    time.Time         `datastore:"create_date,noindex"`
	UpdateDate    time.Time         `datastore:"update_date,noindex"`
	DeleteFlag    bool              `datastore:"delete_flag"`
	DeleteDate    time.Time         `datastore:"delete_date,noindex"`
	Contents      []content.Content `datastore:"contents,noindex"`
}

// Site はサイト情報の取得結果を示す
type Site struct {
	FeedURL       string `json:"FeedUrl"`
	Type          string
	SiteURL       string `json:"SiteUrl"`
	SiteTitle     string
	SiteIcon      utility.ExURL
	LatestContent Content
	HubURL        string `json:"HubUrl"`
	Contents      []content.Content
	CreateDate    time.Time
}

// Content コンテンツ情報
type Content struct {
	URL   string        `datastore:"url,noindex"`
	Title string        `datastore:"title,noindex"`
	Image utility.ExURL `datastore:"image,noindex"`
}

func (s *physicalSite) createSecret() string {
	return s.CreateDate.Format("20060102031605")
}

// site はphysicalSiteからサイト情報(Site)への変換を行う
func (s physicalSite) site() *Site {
	return &Site{
		FeedURL:       s.Key,
		Type:          s.Type,
		SiteURL:       s.SiteURL,
		SiteTitle:     s.SiteTitle,
		SiteIcon:      utility.ExURL(s.SiteIcon),
		LatestContent: s.LatestContent,
		HubURL:        s.HubURL,
		Contents:      s.Contents,
		CreateDate:    s.CreateDate,
	}
}

// FromFeedURL はFeedURLを元にSiteを取得する
func FromFeedURL(ctx context.Context, feedURL string) (*Site, error) {
	g := goon.FromContext(ctx)
	s := &physicalSite{Key: feedURL}
	err := g.Get(s)
	if err != nil {
		return nil, err
	}
	return s.site(), nil
}

// DeleteUnnecessarySite は三ヶ月以上更新がないサイトを抽出し削除する
func DeleteUnnecessarySite(ctx context.Context) error {
	g := goon.FromContext(ctx)
	query := datastore.NewQuery("physicalSite").Filter("UpdateDate <=", time.Now().AddDate(0, -3, 0)).KeysOnly()

	keys, err := g.GetAll(query, nil)
	if err != nil {
		return errors.New("DeleteUnnecessarySite: g.GetAll: " + err.Error())
	}
	err = g.DeleteMulti(keys)
	if err != nil {
		return errors.New("DeleteUnnecessarySite: g.DeleteMulti: " + err.Error())
	}
	return nil
}

func getBodyByURL(ctx context.Context, url string) ([]byte, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		log.Infof(ctx, "client.Get error url:%v, err:%v", url, err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Infof(ctx, "http.Status NG url %s, resp %v", url, resp)
		return nil, errors.New("unknown status code")
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func getContentsInfo(ctx context.Context, url string) (*Site, error) {
	body, err := getBodyByURL(ctx, url)
	if err != nil {
		return nil, err
	}
	feed, err := getFeedInfo(ctx, body)
	if err == nil {
		// feed解析成功
		h, _ := utility.ParseHTML(ctx, feed.SiteURL)
		c := content.Convert(ctx, feed.Contents)
		result := Site{
			FeedURL:   url,
			Type:      feed.Type,
			SiteURL:   feed.SiteURL,
			SiteTitle: feed.SiteTitle,
			SiteIcon:  utility.ExURL(h.IconURL),
			LatestContent: Content{
				URL:   c[0].URL,
				Title: c[0].Title,
				Image: c[0].ImageURL,
			},
			Contents:   c,
			HubURL:     feed.HubURL,
			CreateDate: time.Now(),
		}
		return &result, nil
	}

	// html形式か？
	h, err := utility.ParseHTML(ctx, url)
	if err == nil && h.FeedURL != "" {
		return getContentsInfo(ctx, h.FeedURL)
	}

	// 過去の経緯から[url]atom.xmlをチェックする
	if len(url)-1 == strings.LastIndex(url, "/") {
		return getContentsInfo(ctx, url+"atom.xml")
	}

	return nil, errors.New("can't read information")
}

func getFeedInfo(ctx context.Context, body []byte) (*content.Feed, error) {
	// 入力データをクリーニング
	body = []byte(strings.Replace(string(body), string('\u000c'), "", -1))

	// ATOM形式か?
	feed, err := atom.Analyze(body)
	if err == nil {
		return &feed, nil
	}

	// RSS 1.0形式か?
	feed, err = rdf.Analyze(body)
	if err == nil {
		log.Debugf(ctx, "Feed(RDF) %v", feed)
		return &feed, nil
	}

	// RSS 2.0形式か?
	feed, err = rss.Analyze(body)
	if err == nil {
		return &feed, nil
	}

	return nil, errors.New("not feed")
}
