package rdf

import (
	"reflect"
	"testing"
	"time"

	"github.com/info-matopush/matopush/src/site"
)

func TestRead(t *testing.T) {
	data := `<rdf>
	<channel>
	<title>Title</title>
	<link>http://sample.com</link>
	<atom:link href="http://sample.com/feed" rel="self" type="application/rss+xml" />
	<atom:link rel="hub" href="http://pubsubhubbub.appspot.com"/>
	</channel>
	<item>
		<title>Content Title</title>
		<link>http://sample.com/content.html</link>
		<description><![CDATA[description]]></description>
		<dc:date>2018-03-31T10:02:32+09:00</dc:date>
	</item>
	</rdf>`
	feed, err := Analyze([]byte(data))
	if err != nil {
		t.Fatalf("xml.Unmarshal error: %+v", err)
	}

	expect := site.Feed{
		Type:      "RSS 1.0",
		SiteTitle: "Title",
		SiteURL:   "http://sample.com",
		HubURL:    "http://pubsubhubbub.appspot.com",
		Contents: []site.ContentFromFeed{
			site.ContentFromFeed{
				Title:      "Content Title",
				URL:        "http://sample.com/content.html",
				Summary:    "description",
				ModifyDate: time.Date(2018, time.March, 31, 10, 2, 32, 0, time.Local),
			},
		},
	}

	if !reflect.DeepEqual(feed, expect) {
		t.Fatalf("not match: feed %+v, expect %+v", feed, expect)
	}
	t.Logf("result: %+v", feed)
}
