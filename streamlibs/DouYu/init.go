package DouYu

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Link struct {
	rid        string
	did        string
	t10        string
	t13        string
	apiErrCode int64
	res        string

	debug bool

	client *http.Client
}

type AddHeaderTransport struct {
	T http.RoundTripper
}

func (adt *AddHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.36")
	return adt.T.RoundTrip(req)
}

func NewAddHeaderTransport(T http.RoundTripper) *AddHeaderTransport {
	if T == nil {
		T = http.DefaultTransport
	}
	return &AddHeaderTransport{T}
}

func NewDouyuLink(rid string, proxy string, debug bool) (*Link, error) {
	var (
		err      error
		proxyURL *url.URL
	)
	dy := new(Link)
	dy.t10 = strconv.Itoa(int(time.Now().Unix()))
	dy.t13 = strconv.Itoa(int(time.Now().UnixMilli()))
	dy.debug = debug
	if len(proxy) > 0 {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		dy.client = &http.Client{Transport: NewAddHeaderTransport(&http.Transport{Proxy: http.ProxyURL(proxyURL)})}
	} else {
		dy.client = &http.Client{Transport: NewAddHeaderTransport(nil)}
	}
	dy.rid, err = dy.getRealRoomID(rid)
	if err != nil {
		return nil, err
	}
	dy.did, err = dy.getDeviceID()
	if err != nil {
		return nil, err
	}
	_, err = dy.getPreData()
	if err != nil {
		return nil, fmt.Errorf("获取直播间信息异常：%w", err)
	}
	return dy, nil
}
