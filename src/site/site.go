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
	Type         string
}

type Content struct {
	Url   string `datastore:"url,noindex"`
	Title string `datastore:"title,noindex"`
}

// KeyはFeedUrl
type physicalSite struct {
	Key           string    `datastore:"-" goon:"id"`
	Type          string    `datastore:"type,noindex"`
	SiteTitle     string    `datastore:"site_title,noindex"`
	LatestContent Content   `datastore:"latest,noindex"`
	Public        bool      `datastore:"public"`
	HubUrl        string    `datastore:"hub_url,noindex"`
	ContentList   []Content `datastore:"content,noindex"`
	Count         int64     `datastore:"count,noindex"`
	CreateDate    time.Time `datastore:"create_date,noindex"`
	UpdateDate    time.Time `datastore:"update_date,noindex"`
	DeleteFlag    bool      `datastore:"delete_flag"`
	DeleteDate    time.Time `datastore:"delete_date,noindex"`
}

// サイト更新情報
type UpdateInfo struct {
	FeedUrl      string
	SiteTitle    string
	ContentUrl   string
	ContentTitle string
	UpdateFlg    bool
	Icon         string
	Value        bool
	Endpoint     string
	Count        int64
	HubUrl       string
}

func (sui *UpdateInfo) UpdateCount(ctx context.Context, count int64) {
	g := goon.FromContext(ctx)

	s := &physicalSite{Key: sui.FeedUrl}
	err := g.Get(s)
	if err != nil {
		return
	}
	s.Count = count
	g.Put(s)
}

// 三ヶ月以上更新がないサイトを抽出し削除する
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

func List(ctx context.Context) ([]UpdateInfo, error) {
	g := goon.FromContext(ctx)

	var sui []UpdateInfo
	var list []physicalSite
	query := datastore.NewQuery("physicalSite").Filter("delete_flag=", false)
	_, err := g.GetAll(query, &list)
	if err != nil {
		return sui, nil
	}
	for _, s := range list {
		sui = append(sui, UpdateInfo{
			FeedUrl:      s.Key,
			SiteTitle:    s.SiteTitle,
			ContentUrl:   s.LatestContent.Url,
			ContentTitle: s.LatestContent.Title,
		})
	}
	return sui, nil
}

func PublicList(ctx context.Context) ([]UpdateInfo, error) {
	g := goon.FromContext(ctx)

	var sui []UpdateInfo
	var list []physicalSite
	query := datastore.NewQuery("physicalSite").Filter("delete_flag=", false).Filter("public=", true)
	_, err := g.GetAll(query, &list)
	if err != nil {
		return sui, nil
	}
	for _, s := range list {
		sui = append(sui, UpdateInfo{
			FeedUrl:      s.Key,
			SiteTitle:    s.SiteTitle,
			ContentUrl:   s.LatestContent.Url,
			ContentTitle: s.LatestContent.Title,
		})
	}
	log.Infof(ctx, "func PublicList count %v", len(list))
	return sui, nil
}

func (sui *UpdateInfo) Update(ctx context.Context) {
	g := goon.FromContext(ctx)
	s := &physicalSite{Key: sui.FeedUrl}
	g.Get(s)
	s.LatestContent.Url = sui.ContentUrl
	s.LatestContent.Title = sui.ContentTitle
	s.Count = sui.Count
	s.UpdateDate = time.Now()
	g.Put(s)
}

func FromUrl(ctx context.Context, url string) (*UpdateInfo, bool, error) {
	g := goon.FromContext(ctx)
	s := &physicalSite{Key: url}
	err := g.Get(s)
	if err == nil {
		return &UpdateInfo{
			FeedUrl:      s.Key,
			SiteTitle:    s.SiteTitle,
			ContentUrl:   s.LatestContent.Url,
			ContentTitle: s.LatestContent.Title,
			HubUrl:       s.HubUrl,
		}, false, nil
	}
	// 未登録と見做す
	info, err := getContentsInfo(ctx, url)
	if err != nil {
		// 初回読み込み失敗はエラーとみなす
		return nil, false, err
	}
	s.Key = info.FeedUrl
	s.Type = info.Type
	s.SiteTitle = info.SiteTitle
	s.LatestContent.Url = info.ContentUrl
	s.LatestContent.Title = info.ContentTitle
	s.HubUrl = info.HubUrl
	s.CreateDate = time.Now()
	s.UpdateDate = time.Now()
	g.Put(s)
	return &UpdateInfo{
		FeedUrl:      s.Key,
		SiteTitle:    s.SiteTitle,
		ContentUrl:   s.LatestContent.Url,
		ContentTitle: s.LatestContent.Title,
		HubUrl:       s.HubUrl,
	}, true, nil
}

func getBodyByUrl(ctx context.Context, url string) ([]byte, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		log.Infof(ctx, "get error %v, %v", url, err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Infof(ctx, "url %s, resp %v", url, resp)
		return nil, errors.New("unknown status code")
	}
	return ioutil.ReadAll(resp.Body)
}

func (sui *UpdateInfo) CheckSite(ctx context.Context) error {
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
	}
	return nil
}

func CheckSiteByFeed(ctx context.Context, url string, body []byte) (*UpdateInfo, error) {
	sui, _, err := FromUrl(ctx, url)
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
		sui.ContentUrl = info.ContentUrl
		sui.ContentTitle = info.ContentTitle
		sui.UpdateFlg = true
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

	return nil, errors.New("can't read information")
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
			for _, link := range feed.Link {
				if link.Rel == "hub" {
					result.HasHub = true
					result.HubUrl = link.Href
				}
			}
			result.ContentTitle = feed.Entry[0].Title
			result.Type = "atom"
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
			result.Type = "rss1.0"
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

			if rss2.Channel.AtomLink.Rel == "hub" {
				result.HasHub = true
				result.HubUrl = rss1.Channel.AtomLink.Href
			}
			result.Type = "rss2.0"
			return &result, nil
		}
	}
	return nil, errors.New("not feed")
}
