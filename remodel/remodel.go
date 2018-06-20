package remodel

import (
	"encoding/json"
	"strings"
)

// ExURL は拡張URL型
type ExURL string

// MarshalJSON は拡張ExURLをJSON出力する
func (e ExURL) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.TunneledURL())
}

// TunneledURL はproxyされたURLを返却する
func (e ExURL) TunneledURL() string {
	if strings.HasPrefix(string(e), "http://") {
		return "/api/tunnel?url=" + string(e)
	}
	return string(e)
}
