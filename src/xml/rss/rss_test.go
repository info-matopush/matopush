package rss

import (
	"encoding/xml"
	"testing"
)

func TestRead(t *testing.T) {
	s := `<rss>
	<channel>
	<title>Title</title>
	<link>http://sample.com</link>
	<atom:link href="http://sample.com/feed" rel="self" type="application/rss+xml" />
	<item>
		<title>Content Title</title>
		<link>http://sample.com/content.html</link>
		<description><![CDATA[description]]>
		</description>
	</item>
	</channel>
	</rss>`
	rss := RSS{}
	err := xml.Unmarshal([]byte(s), &rss)
	if err != nil {
		t.Fatalf("xml.Unmarshal error: %v", err)
	}
	t.Logf("result: %+v", rss)
}
