package DouYu

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/nv4d1k/streamlink-go/app/lib"
)

type Link struct {
	rid        string
	did        string
	t10        string
	t13        string
	apiErrCode int64
	res        string
	proxy      *url.URL

	debug bool

	client *http.Client
}

func NewDouyuLink(rid string, proxy *url.URL, debug bool) (*Link, error) {
	var (
		err      error
		proxyURL *url.URL
	)
	dy := new(Link)
	dy.t10 = strconv.Itoa(int(time.Now().Unix()))
	dy.t13 = strconv.Itoa(int(time.Now().UnixMilli()))
	dy.debug = debug
	if proxy != nil {
		dy.client = &http.Client{Transport: lib.NewAddHeaderTransport(&http.Transport{Proxy: http.ProxyURL(proxyURL)}, false)}
		dy.proxy = proxyURL
	} else {
		dy.client = &http.Client{Transport: lib.NewAddHeaderTransport(nil, false)}
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
		return nil, fmt.Errorf("get pre data error：%w", err)
	}
	return dy, nil
}
