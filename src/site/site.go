package site

import (
	"encoding/xml"
	"errors"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
	"io/ioutil"
	"net/http"
	"src/atom"
	"time"
	"google.golang.org/appengine/datastore"
	"src/rdf"
	"src/html"
	"strings"
	"src/rss"
)

type Result struct {
	SiteTitle    string
	ContentTitle string
	ContentUrl   string
}

type SiteUpdateInfo struct {
	SiteUrl      string    `datastore:"-"                   goon:"id"`
	SiteTitle    string    `datastore:"site_title,noindex"`
	ContentUrl   string    `datastore:"content_url,noindex"`
	ContentTitle string    `datastore:"content_title,noindex"`
	UpdateFlg    bool      `datastore:"-"                   json:"-"`
	Public       bool      `datastore:"public"`
	Icon         string    `datastore:"-"`
	CreateDate   time.Time `datastore:"create_date,noindex" json:"-"`
	UpdateDate   time.Time `datastore:"update_date,noindex" json:"-"`
	DeleteFlag   bool      `datastore:"delete_flag"         json:"-"`
	DeleteDate   time.Time `datastore:"delete_date,noindex" json:"-"`
	// マイ・リストを返す時だけ使う
	Value        string    `datastore:"-"`
	// プッシュ通知を行う時だけ使う
	Endpoint     string    `datastore:"-"`
	// cron実行時に通知したEndpointの数を設定する
	SubscribeCount int64     `datastore:"SubscribeCount"`
}

func GetAll(ctx context.Context, dst *[]SiteUpdateInfo) (error) {
	g := goon.FromContext(ctx)
	query := datastore.NewQuery("SiteUpdateInfo").Filter("delete_flag=", false)
	keys, err := g.GetAll(query, dst)
	log.Infof(ctx, "keys num. %d", len(keys))
	return err
}

func Get(ctx context.Context, url string) (*SiteUpdateInfo, error) {
	g := goon.FromContext(ctx)
	sui := &SiteUpdateInfo{SiteUrl:url}
	err := g.Get(sui)
	if err != nil {
		// 未登録と見做す
		info, err := getContentsInfo(ctx, url)
		if err != nil {
			// 初回読み込み失敗はエラーとみなす
			return nil, err
		}

		sui.SiteTitle = info.SiteTitle
		sui.ContentUrl = info.ContentUrl
		sui.ContentTitle = info.ContentTitle
		sui.UpdateFlg = true
		sui.CreateDate = time.Now()
		sui.UpdateDate = time.Now()
		g.Put(sui)
		return sui, nil
	}

	err = CheckSite(ctx, sui)
	return sui, err
}

func CheckSite(ctx context.Context, sui *SiteUpdateInfo) (error) {
	g := goon.FromContext(ctx)

	info, err := getContentsInfo(ctx, sui.SiteUrl)
	if err != nil {
		// Feedの読み込みに失敗
		// todo: どうしよう。とりあえず更新なしとする
		log.Warningf(ctx, "feedの読み込みに失敗 url:%s", sui.SiteUrl)
		return nil
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
	return nil
}

func getContentsInfo(ctx context.Context, url string) (*Result, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		log.Infof(ctx, "site get error. %v", err)
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Infof(ctx, "url %s, resp %v", url, resp)
		return nil, errors.New("Unknows status code.")
	}

	body, _ := ioutil.ReadAll(resp.Body)

	result := Result{}

	// atom.xmlか?
	feed := atom.Feed{}
	err = xml.Unmarshal(body, &feed)
	if err == nil {
		log.Infof(ctx, "atom形式で解析:%s, %v", url, feed)
		result.SiteTitle = feed.Title
		if len(feed.Entry) > 0 {
			// 最初のエントリを返す
			// linkの中の"alternate"を探す
			for _, link := range feed.Entry[0].Link {
				if link.Rel == "alternate" {
					result.ContentUrl = link.Href
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
		log.Infof(ctx, "RSS v1.0形式で解析:%s, %v", url, rss1)
		result.SiteTitle = rss1.Channel.Title
		if len(rss1.Item) > 0 {
			result.ContentTitle = rss1.Item[0].Title
			result.ContentUrl = rss1.Item[0].Link
			return &result, nil
		}
	}

	// RSS2.0か?
	rss2 := rss.RSS{}
	err = xml.Unmarshal(body, &rss2)
	if err == nil {
		log.Infof(ctx, "RSS v2.0形式で解析:%s, %v", url, rss2)
		result.SiteTitle = rss2.Channel.Title
		if len(rss2.Channel.Item) > 0 {
			result.ContentTitle = rss2.Channel.Item[0].Title
			result.ContentUrl = rss2.Channel.Item[0].Link
			return &result, nil
		}

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
	if len(url) - 1 == strings.LastIndex(url, "/") {
		return getContentsInfo(ctx, url + "atom.xml")
	}

	return nil, errors.New("can't read information.")
}
