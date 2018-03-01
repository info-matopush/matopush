package site

import (
	"encoding/xml"
	"errors"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/http"
	"src/xml/atom"
	"src/xml/html"
	"src/xml/rdf"
	"src/xml/rss"
	"strings"
	"time"
)

type Result struct {
	SiteTitle    string
	ContentTitle string
	ContentUrl   string
	FeedUrl      string
	HasHub       bool
	HubUrl       string
}

// サイト更新情報
type SiteUpdateInfo struct {
	SiteUrl        string    `datastore:"-"                   goon:"id"`
	FeedUrl        string    `datastore:"feed_url,noindex"`
	SiteTitle      string    `datastore:"site_title,noindex"`
	ContentUrl     string    `datastore:"content_url,noindex"`
	ContentTitle   string    `datastore:"content_title,noindex"`
	UpdateFlg      bool      `datastore:"-"                   json:"-"`
	Public         bool      `datastore:"public"`
	Icon           string    `datastore:"-"`
	CreateDate     time.Time `datastore:"create_date,noindex" json:"-"`
	UpdateDate     time.Time `datastore:"update_date"         json:"-"`
	DeleteFlag     bool      `datastore:"delete_flag"         json:"-"`
	DeleteDate     time.Time `datastore:"delete_date,noindex" json:"-"`
	Value          string    `datastore:"-"`                      // マイ・リストを返す時だけ使う
	Endpoint       string    `datastore:"-"`                      // プッシュ通知を行う時だけ使う
	SubscribeCount int64     `datastore:"SubscribeCount,noindex"` // cron実行時に通知したEndpointの数を設定する
	HasHub         bool      `datastore:"-"`                      // PubSubHubBubを使用しているか
	HubUrl         string    `datastore:"-"`
}

// 三ヶ月以上更新がないサイトを抽出し削除する
func DeleteUnnecessarySite(ctx context.Context) error {
	g := goon.FromContext(ctx)
	query := datastore.NewQuery("SiteUpdateInfo").Filter("UpdateDate <=", time.Now().AddDate(0, -3, 0)).KeysOnly()

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

func GetAll(ctx context.Context, dst *[]SiteUpdateInfo) error {
	g := goon.FromContext(ctx)
	query := datastore.NewQuery("SiteUpdateInfo").Filter("delete_flag=", false)
	keys, err := g.GetAll(query, dst)
	log.Infof(ctx, "keys num. %d", len(keys))
	return err
}

func Get(ctx context.Context, url string) (*SiteUpdateInfo, bool, error) {
	g := goon.FromContext(ctx)
	sui := &SiteUpdateInfo{SiteUrl: url}
	err := g.Get(sui)
	if err == nil {
		return sui, false, nil
	}
	// 未登録と見做す
	info, err := getContentsInfo(ctx, url)
	if err != nil {
		// 初回読み込み失敗はエラーとみなす
		return nil, false, err
	}

	sui.SiteTitle = info.SiteTitle
	sui.FeedUrl = info.FeedUrl
	sui.ContentUrl = info.ContentUrl
	sui.ContentTitle = info.ContentTitle
	sui.UpdateFlg = true
	sui.CreateDate = time.Now()
	sui.UpdateDate = time.Now()
	sui.HasHub = info.HasHub
	sui.HubUrl = info.HubUrl
	return sui, true, nil
}

func getBodyByUrl(ctx context.Context, url string) ([]byte, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		log.Infof(ctx, "get error. %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Infof(ctx, "url %s, resp %v", url, resp)
		return nil, errors.New("Unknows status code.")
	}
	return ioutil.ReadAll(resp.Body)
}

func CheckSite(ctx context.Context, sui *SiteUpdateInfo) error {
	if sui.FeedUrl == "" {
		info, err := getContentsInfo(ctx, sui.SiteUrl)
		if err != nil {
			return err
		}
		sui.FeedUrl = info.FeedUrl

		// 読み込んだ情報を前回値と比較する
		if sui.ContentUrl != info.ContentUrl {
			sui.SiteTitle = info.SiteTitle
			sui.ContentUrl = info.ContentUrl
			sui.ContentTitle = info.ContentTitle
			sui.UpdateFlg = true
			sui.UpdateDate = time.Now()
		}
	} else {
		body, err := getBodyByUrl(ctx, sui.FeedUrl)
		if err != nil {
			return err
		}
		info, err := getFeedInfo(ctx, body)
		if err != nil {
			return err
		}
		// 読み込んだ情報を前回値と比較する
		if sui.ContentUrl != info.ContentUrl {
			sui.SiteTitle = info.SiteTitle
			sui.ContentUrl = info.ContentUrl
			sui.ContentTitle = info.ContentTitle
			sui.UpdateFlg = true
			sui.UpdateDate = time.Now()
		}
	}
	return nil
}

func CheckSiteByFeed(ctx context.Context, url string, body []byte) (*SiteUpdateInfo, error) {
	g := goon.FromContext(ctx)
	sui := &SiteUpdateInfo{SiteUrl: url}
	err := g.Get(sui)
	if err != nil {
		return nil, err
	}
	// bodyの内容がfeedか判定する
	info, err := getFeedInfo(ctx, body)
	if err != nil {
		return nil, err
	}
	// 読み込んだ情報を前回値と比較する
	if sui.ContentUrl != info.ContentUrl {
		sui.SiteTitle = info.SiteTitle
		sui.ContentUrl = info.ContentUrl
		sui.ContentTitle = info.ContentTitle
		sui.UpdateFlg = true
		sui.UpdateDate = time.Now()
		g.Put(sui)
	}
	return sui, nil
}

func getContentsInfo(ctx context.Context, url string) (*Result, error) {
	body, err := getBodyByUrl(ctx, url)
	if err != nil {
		return nil, err
	}
	result, err := getFeedInfo(ctx, body)
	if err == nil {
		// feed解析成功
		result.FeedUrl = url
		return result, nil
	}

	// html形式か？
	h := html.Html{}
	err = xml.Unmarshal(body, &h)
	if err != nil {
		log.Infof(ctx, "html形式で解析:%s, %v", url, h)
		// head -> linkの中からRSSもしくはAtomを探す
		for _, link := range h.Head.Link {
			log.Infof(ctx, "Link:%v", link)
			if link.Type == "application/rss+xml" {
				return getContentsInfo(ctx, link.Href)
			} else if link.Type == "application/atom+xml" {
				return getContentsInfo(ctx, link.Href)
			}
		}
	}
	// 過去の経緯から[url]atom.xmlをチェックする
	if len(url)-1 == strings.LastIndex(url, "/") {
		return getContentsInfo(ctx, url+"atom.xml")
	}

	return nil, errors.New("can't read information.")
}

func getFeedInfo(ctx context.Context, body []byte) (*Result, error) {
	result := Result{HasHub: false}

	// atom.xmlか?
	feed := atom.Feed{}
	err := xml.Unmarshal(body, &feed)
	if err == nil {
		log.Infof(ctx, "atom形式で解析:%v", feed)
		result.SiteTitle = feed.Title
		if len(feed.Entry) > 0 {
			// 最初のエントリを返す
			// linkの中の"alternate"を探す
			for _, link := range feed.Entry[0].Link {
				if link.Rel == "alternate" {
					result.ContentUrl = link.Href
				}
			}
			// linkの中の"alternate"を探す
			for _, link := range feed.Entry[0].Link {
				if link.Rel == "hub" {
					result.HasHub = true
					result.HubUrl = link.Href
				}
			}
			result.ContentTitle = feed.Entry[0].Title
			return &result, nil
		}
	}

	// atom.xmlでないならばRSS1.0か?
	rss1 := rdf.RDF{}
	err = xml.Unmarshal(body, &rss1)
	if err == nil {
		log.Infof(ctx, "RSS v1.0形式で解析:%v", rss1)
		result.SiteTitle = rss1.Channel.Title
		if len(rss1.Item) > 0 {
			result.ContentTitle = rss1.Item[0].Title
			result.ContentUrl = rss1.Item[0].Link
			if rss1.Channel.AtomLink.Rel == "hub" {
				result.HasHub = true
				result.HubUrl = rss1.Channel.AtomLink.Href
			}
			return &result, nil
		}
	}

	// RSS2.0か?
	rss2 := rss.RSS{}
	err = xml.Unmarshal(body, &rss2)
	if err == nil {
		log.Infof(ctx, "RSS v2.0形式で解析:%v", rss2)
		result.SiteTitle = rss2.Channel.Title
		if len(rss2.Channel.Item) > 0 {
			result.ContentTitle = rss2.Channel.Item[0].Title
			result.ContentUrl = rss2.Channel.Item[0].Link
			return &result, nil
		}

	}
	return nil, errors.New("not feed.")
}
