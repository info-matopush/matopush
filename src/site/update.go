package site

import (
	"time"

	"github.com/info-matopush/matopush/src/content"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
)

// UpdateInfo はサイト更新情報
type UpdateInfo struct {
	Site
	UpdateFlg bool
	Value     bool
	Count     int64
	Secret    string // pubsubhubbubで使用する秘密鍵
}

func fromPhysicalSite(s physicalSite) UpdateInfo {
	return UpdateInfo{
		Site:      s.Site(),
		UpdateFlg: false,
		Value:     false,
		Count:     0,
		Secret:    s.createSecret(),
	}
}

// List は全サイト情報を取得する
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

// PublicList は公開サイトリストを取得する
func PublicList(ctx context.Context) (dst []UpdateInfo) {
	g := goon.FromContext(ctx)

	var list []physicalSite
	query := datastore.NewQuery("physicalSite").Filter("delete_flag=", false).Filter("public=", true)
	_, err := g.GetAll(query, &list)
	if err != nil {
		log.Errorf(ctx, "g.GetAll error %v", err)
		return
	}
	for _, s := range list {
		dst = append(dst, fromPhysicalSite(s))
	}
	log.Infof(ctx, "func PublicList count %v", len(list))
	return
}

// Update はサイト情報を更新する
func (ui *UpdateInfo) Update(ctx context.Context) {
	g := goon.FromContext(ctx)
	s := &physicalSite{Key: ui.FeedURL}
	g.Get(s)
	s.Type = ui.Type
	s.SiteURL = ui.SiteURL
	s.SiteTitle = ui.SiteTitle
	s.SiteIcon = string(ui.SiteIcon)
	s.LatestContent = ui.LatestContent
	s.Count = ui.Count
	s.Contents = ui.Contents
	s.UpdateDate = time.Now()
	g.Put(s)
}

// CheckSite はサイトにアクセスして更新情報を取得する
func (ui *UpdateInfo) CheckSite(ctx context.Context) error {
	body, err := getBodyByURL(ctx, ui.FeedURL)
	if err != nil {
		return err
	}
	feed, err := getFeedInfo(ctx, body)
	if err != nil {
		return err
	}
	// 読み込んだ情報を前回値と比較する
	ui.Type = feed.Type
	ui.SiteURL = feed.SiteURL
	ui.SiteTitle = feed.SiteTitle
	ui.Contents = content.Convert(ctx, feed.Contents)
	ui.HubURL = feed.HubURL
	if ui.LatestContent.URL != feed.Contents[0].URL {
		ui.LatestContent.URL = feed.Contents[0].URL
		ui.LatestContent.Title = feed.Contents[0].Title
		ui.UpdateFlg = true
	}
	return nil
}

// FromURL はURLから更新情報を作成する
func FromURL(ctx context.Context, url string) (*UpdateInfo, bool, error) {
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
	s.SiteIcon = string(info.SiteIcon)
	s.LatestContent = info.LatestContent
	s.Contents = info.Contents
	s.HubURL = info.HubURL
	s.CreateDate = time.Now()
	s.UpdateDate = time.Now()
	g.Put(&s)
	ui := fromPhysicalSite(s)
	return &ui, true, nil
}

// CheckSiteByFeed はフィード内容から更新情報を出力する
func CheckSiteByFeed(ctx context.Context, url string, body []byte) (*UpdateInfo, error) {
	ui, _, err := FromURL(ctx, url)
	if err != nil {
		return nil, err
	}
	// bodyの内容がfeedか判定する
	feed, err := getFeedInfo(ctx, body)
	if err != nil {
		return nil, err
	}
	// 読み込んだ情報を前回値と比較する
	ui.Type = feed.Type
	ui.SiteURL = feed.SiteURL
	ui.SiteTitle = feed.SiteTitle
	ui.Contents = content.Convert(ctx, feed.Contents)
	ui.HubURL = feed.HubURL
	if ui.LatestContent.URL != feed.Contents[0].URL {
		ui.LatestContent.URL = feed.Contents[0].URL
		ui.LatestContent.Title = feed.Contents[0].Title
		ui.UpdateFlg = true
	}
	return ui, nil
}
