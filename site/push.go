package site

import (
	"github.com/info-matopush/matopush/utility"
	"github.com/mjibson/goon"
	"golang.org/x/net/context"
	"google.golang.org/appengine/log"
)

// ForPush はプッシュ通知用の限定された情報
type ForPush struct {
	FeedURL      string
	SiteURL      string
	SiteTitle    string
	SiteIcon     utility.ExURL
	ContentURL   string
	ContentTitle string
	ContentImage utility.ExURL
}

// GetSiteForPush はプッシュ通知用のサイト情報を取得する
func GetSiteForPush(ctx context.Context, feedURL string) (push ForPush, err error) {
	g := goon.FromContext(ctx)
	s := &physicalSite{Key: feedURL}
	err = g.Get(s)
	if err != nil {
		log.Errorf(ctx, "GetSiteForPush:Get error : %s", feedURL)
		return
	}
	push.FeedURL = feedURL
	push.SiteURL = s.SiteURL
	push.SiteTitle = s.SiteTitle
	push.SiteIcon = utility.ExURL(s.SiteIcon)
	push.ContentURL = s.LatestContent.URL
	push.ContentTitle = s.LatestContent.Title
	push.ContentImage = s.LatestContent.Image
	return
}
