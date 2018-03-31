package rdf

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"testing"
)

func TestRead(t *testing.T) {
	resp, err := http.Get("http://hosyusokuhou.jp/feed")
	if err != nil {
		t.Fatalf("http.Get error: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("resp.StatusCode error: %v", resp.StatusCode)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("ioutil.ReadAll error: %v", err)
	}
	feed := RDF{}
	err = xml.Unmarshal(body, &feed)
	if err != nil {
		t.Fatalf("xml.Unmarshal error: %v", err)
	}
	t.Logf("result: %+v", feed)
}
