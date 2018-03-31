package site

import (
	"encoding/xml"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/info-matopush/matopush/src/content"
	"github.com/info-matopush/matopush/src/xml/atom"
	"github.com/info-matopush/matopush/src/xml/rdf"
	"github.com/info-matopush/matopush/src/xml/rss"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"
)

type Result struct {
	SiteURL      string
	SiteTitle    string
	ContentTitle string
	ContentURL   string
	FeedURL      string
	HasHub       bool
	HubURL       string
	Type         string
	Contents     []content.Content
}

type Content struct {
	Title string `datastore:"title,noindex"`
	URL   string `datastore:"url,noindex"`
}

// KeyはFeedUrl
type physicalSite struct {
	Key           string            `datastore:"-" goon:"id"`
	Type          string            `datastore:"type,noindex"`
	SiteURL       string            `datastore:"site_url,noindex"`
	SiteTitle     string            `datastore:"site_title,noindex"`
	LatestContent Content           `datastore:"latest,noindex"`
	Public        bool              `datastore:"public"`
	HubURL        string            `datastore:"hub_url,noindex"`
	ContentList   []Content         `datastore:"content,noindex"`
	Count         int64             `datastore:"count,noindex"`
	CreateDate    time.Time         `datastore:"create_date,noindex"`
	UpdateDate    time.Time         `datastore:"update_date,noindex"`
	DeleteFlag    bool              `datastore:"delete_flag"`
	DeleteDate    time.Time         `datastore:"delete_date,noindex"`
	Contents      []content.Content `datastore:"contents,noindex"`
}

// サイト更新情報
type UpdateInfo struct {
	FeedURL      string `json:"FeedUrl"`
	SiteURL      string `json:"SiteUrl"`
	SiteTitle    string
	ContentURL   string `json:"ContentUrl"`
	ContentTitle string
	UpdateFlg    bool
	Icon         string
	Value        bool
	Endpoint     string
	Count        int64
	HubURL       string `json:"HubUrl"`
	Secret       string // pubsubhubbubで使用する秘密鍵
	Type         string
	Contents     []content.Content
}

func (s *physicalSite) createSecret() string {
	return s.CreateDate.Format("20060102031605")
}

func (ui *UpdateInfo) UpdateCount(ctx context.Context, count int64) {
	g := goon.FromContext(ctx)

	s := &physicalSite{Key: ui.FeedURL}
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

func fromPhysicalSite(s physicalSite) UpdateInfo {
	return UpdateInfo{
		FeedURL:      s.Key,
		SiteURL:      s.SiteURL,
		SiteTitle:    s.SiteTitle,
		ContentURL:   s.LatestContent.URL,
		ContentTitle: s.LatestContent.Title,
		UpdateFlg:    false,
		Icon:         "",
		Value:        false,
		Endpoint:     "",
		Count:        0,
		HubURL:       s.HubURL,
		Secret:       s.createSecret(),
		Type:         s.Type,
		Contents:     s.Contents,
	}
}

func List(ctx context.Context) ([]UpdateInfo, error) {
	g := goon.FromContext(ctx)

	var ui []UpdateInfo
	var list []physicalSite
	query := datastore.NewQuery("physicalSite").Filter("delete_flag=", false)
	_, err := g.GetAll(query, &list)
	if err != nil {
		return ui, nil
	}
	for _, s := range list {
		ui = append(ui, fromPhysicalSite(s))
	}
	return ui, nil
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
		sui = append(sui, fromPhysicalSite(s))
	}
	log.Infof(ctx, "func PublicList count %v", len(list))
	return sui, nil
}

func (ui *UpdateInfo) Update(ctx context.Context) {
	g := goon.FromContext(ctx)
	s := &physicalSite{Key: ui.FeedURL}
	g.Get(s)
	s.LatestContent.URL = ui.ContentURL
	s.LatestContent.Title = ui.ContentTitle
	s.Count = ui.Count
	s.UpdateDate = time.Now()
	s.Contents = ui.Contents
	g.Put(s)
}

func FromUrl(ctx context.Context, url string) (*UpdateInfo, bool, error) {
	g := goon.FromContext(ctx)
	s := physicalSite{Key: url}
	err := g.Get(&s)
	if err == nil {
		ui := fromPhysicalSite(s)
		return &ui, false, nil
	}
	// 未登録と見做す
	info, err := getContentsInfo(ctx, url)
	if err != nil {
		// 初回読み込み失敗はエラーとみなす
		return nil, false, err
	}
	s.Key = info.FeedURL
	s.Type = info.Type
	s.SiteURL = info.SiteURL
	s.SiteTitle = info.SiteTitle
	s.LatestContent.URL = info.ContentURL
	s.LatestContent.Title = info.ContentTitle
	s.Contents = info.Contents
	s.HubURL = info.HubURL
	s.CreateDate = time.Now()
	s.UpdateDate = time.Now()
	g.Put(&s)
	ui := fromPhysicalSite(s)
	return &ui, true, nil
}

func getBodyByUrl(ctx context.Context, url string) ([]byte, error) {
	client := urlfetch.Client(ctx)
	resp, err := client.Get(url)
	if err != nil {
		log.Infof(ctx, "get error %v, %v", url, err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		log.Infof(ctx, "url %s, resp %v", url, resp)
		return nil, errors.New("unknown status code")
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (ui *UpdateInfo) CheckSite(ctx context.Context) error {
	body, err := getBodyByUrl(ctx, ui.FeedURL)
	if err != nil {
		return err
	}
	info, err := getFeedInfo(ctx, body)
	if err != nil {
		return err
	}
	// 読み込んだ情報を前回値と比較する
	if ui.ContentURL != info.ContentURL {
		ui.SiteURL = info.SiteURL
		ui.SiteTitle = info.SiteTitle
		ui.ContentTitle = info.ContentTitle
		ui.ContentURL = info.ContentURL
		ui.Contents = info.Contents
		ui.UpdateFlg = true
	}
	return nil
}

func CheckSiteByFeed(ctx context.Context, url string, body []byte) (*UpdateInfo, error) {
	ui, _, err := FromUrl(ctx, url)
	if err != nil {
		return nil, err
	}
	// bodyの内容がfeedか判定する
	info, err := getFeedInfo(ctx, body)
	if err != nil {
		return nil, err
	}
	// 読み込んだ情報を前回値と比較する
	if ui.ContentURL != info.ContentURL {
		ui.SiteURL = info.SiteURL
		ui.SiteTitle = info.SiteTitle
		ui.ContentURL = info.ContentURL
		ui.ContentTitle = info.ContentTitle
		ui.Contents = info.Contents
		ui.UpdateFlg = true
	}
	return ui, nil
}

func getContentsInfo(ctx context.Context, url string) (*Result, error) {
	body, err := getBodyByUrl(ctx, url)
	if err != nil {
		return nil, err
	}
	result, err := getFeedInfo(ctx, body)
	if err == nil {
		// feed解析成功
		result.FeedURL = url
		return result, nil
	}

	// html形式か？
	h, err := content.HtmlParse(ctx, url)
	if err == nil && h.FeedURL != "" {
		return getContentsInfo(ctx, h.FeedURL)
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
					result.ContentURL = link.Href
				}
			}
			for _, link := range feed.Link {
				if link.Rel == "alternate" {
					result.SiteURL = link.Href
				} else if link.Rel == "hub" {
					result.HasHub = true
					result.HubURL = link.Href
				}
			}
			result.ContentTitle = feed.Entry[0].Title
			result.Type = "atom"

			var c []content.Content
			for _, i := range feed.ListContentFromFeed() {
				con, err := content.New(ctx, i)
				if err == nil {
					c = append(c, *con)
				}
			}
			result.Contents = c

			return &result, nil
		}
	}

	// atom.xmlでないならばRSS1.0か?
	rdf := rdf.RDF{}
	err = xml.Unmarshal(body, &rdf)
	if err == nil {
		log.Infof(ctx, "RSS v1.0形式で解析:%v", rdf)
		result.SiteTitle = rdf.Channel.Title
		if len(rdf.Item) > 0 {
			result.ContentTitle = rdf.Item[0].Title
			result.ContentURL = rdf.Item[0].Link
			for _, l := range rdf.Channel.Link {
				if l.Rel == "hub" {
					result.HasHub = true
					result.HubURL = l.Href
				} else if l.Data != "" {
					result.SiteURL = l.Data
				}
			}
			result.Type = "rss1.0"

			var c []content.Content
			for _, i := range rdf.ListContentFromFeed() {
				con, err := content.New(ctx, i)
				if err == nil {
					c = append(c, *con)
				}
			}
			result.Contents = c

			return &result, nil
		}
	}

	// RSS2.0か?
	rss := rss.RSS{}
	err = xml.Unmarshal(body, &rss)
	if err == nil {
		log.Infof(ctx, "RSS v2.0形式で解析:%v", rss)
		result.SiteTitle = rss.Channel.Title
		if len(rss.Channel.Item) > 0 {
			result.ContentTitle = rss.Channel.Item[0].Title
			result.ContentURL = rss.Channel.Item[0].Link

			for _, l := range rdf.Channel.Link {
				if l.Rel == "hub" {
					result.HasHub = true
					result.HubURL = l.Href
				} else if l.Data != "" {
					result.SiteURL = l.Data
				}
			}

			result.Type = "rss2.0"

			var c []content.Content
			for _, i := range rss.ListContentFromFeed() {
				con, err := content.New(ctx, i)
				if err == nil {
					c = append(c, *con)
				}
			}
			result.Contents = c

			return &result, nil
		}
	}
	return nil, errors.New("not feed")
}
