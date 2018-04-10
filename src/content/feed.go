package content

// Feed はフィード(ATOM, RSS 1.0, RSS 2.0)に含まれる情報
type Feed struct {
	Type      string
	SiteURL   string
	SiteTitle string
	Contents  []FromFeed
	HubURL    string
}
