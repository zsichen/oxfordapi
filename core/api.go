package core

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var URL = func(lang, word string) string {
	return fmt.Sprintf("https://od-api.oxforddictionaries.com/api/v2/entries/%s/%s", lang, strings.ToLower(word))
}

func OxfordAPIRequest(appid, appkey, lang, word string) ([]byte, error) {
	req, err := http.NewRequest(http.MethodGet, URL(lang, word), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("app_id", appid)
	req.Header.Add("app_key", appkey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var buf bytes.Buffer
	io.Copy(&buf, resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return buf.Bytes(), nil
}
