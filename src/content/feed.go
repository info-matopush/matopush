package content

import (
	"time"
)

type Feed struct {
	Type      string
	SiteURL   string
	SiteTitle string
	Contents  []ContentFromFeed
	HubURL    string
}

type ContentFromFeed struct {
	URL        string
	Title      string
	Summary    string
	ModifyDate time.Time
}
