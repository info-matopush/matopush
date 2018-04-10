package content

import (
	"testing"

	"google.golang.org/appengine/aetest"
)

func TestHTMLParse(t *testing.T) {
	ctx, done, _ := aetest.NewContext()
	h, err := ParseHTML(ctx, "https://www.youtube.com/channel/UCu3Mp1ZimtNvyA-bcfo9VrQ")
	t.Logf("HTML %+v, [err %v]", h, err)
	defer done()
}
