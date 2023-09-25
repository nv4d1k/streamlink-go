package HuYa

import (
	"github.com/tidwall/gjson"
	"net/http"
	"net/url"
)

var CDN_TYPE = map[string]string{
	"AL": "阿里", "TX": "腾讯", "HW": "华为", "HS": "火山", "WS": "网宿", "HY": "虎牙",
}

type Link struct {
	rid    string
	uid    string
	uidi   int64
	uuid   string
	res    gjson.Result
	client *http.Client
}

func NewHuyaLink(rid string, proxy string) (*Link, error) {
	var (
		err      error
		proxyURL *url.URL
	)
	hy := new(Link)
	hy.rid = rid
	if len(proxy) > 0 {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		hy.client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	} else {
		hy.client = &http.Client{}
	}
	err = hy.getRoomInfo()
	if err != nil {
		return nil, err
	}
	err = hy.getAnonymousUID()
	if err != nil {
		return nil, err
	}
	hy.getUUID()
	return hy, nil
}
