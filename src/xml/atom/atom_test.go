package atom

import (
	"reflect"
	"testing"
	"time"

	"github.com/info-matopush/matopush/src/site"
)

func TestRead(t *testing.T) {
	data := `<feed>
	<link rel="alternate" href="http://sample.com"/>
	<link rel="hub" href="http://pubsubhubbub.appspot.com"/>
	<title>Title</title>
	<entry>
		<title>Content Title</title>
		<link rel="alternate" href="http://sample.com/content.html"/>
		<summary><![CDATA[description]]></summary>
		<modified>2018-03-31T10:02:32Z</modified>
	</entry>
	</feed>`
	feed, err := Analyze([]byte(data))
	if err != nil {
		t.Fatalf("xml.Unmarshal error: %+v", err)
	}

	expect := site.Feed{
		Type:      "ATOM",
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
