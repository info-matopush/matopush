package rss

import (
	"reflect"
	"testing"
	"time"

	"github.com/info-matopush/matopush/src/content"
)

func TestRead(t *testing.T) {
	data := `<rss>
	<channel>
	<title>Title</title>
	<link>http://sample.com</link>
	<atom:link href="http://sample.com/feed" rel="self" type="application/rss+xml" />
	<item>
		<title>Content Title</title>
		<link>http://sample.com/content.html</link>
		<description><![CDATA[description]]></description>
		<pubDate>Sat, 31 Mar 2018 14:08:32 +0900</pubDate>
		</item>
	</channel>
	</rss>`
	feed, err := Analyze([]byte(data))
	if err != nil {
		t.Fatalf("xml.Unmarshal error: %+v", err)
	}

	expect := content.Feed{
		Type:      "RSS 2.0",
		SiteTitle: "Title",
		SiteURL:   "http://sample.com",
		Contents: []content.FromFeed{
			content.FromFeed{
				Title:      "Content Title",
				URL:        "http://sample.com/content.html",
				Summary:    "description",
				ModifyDate: time.Date(2018, time.March, 31, 14, 8, 32, 0, time.Local),
			},
		},
	}

	if !reflect.DeepEqual(feed, expect) {
		t.Fatalf("not match: feed %+v, expect %+v", feed, expect)
	}
	t.Logf("result: %+v", feed)
}
