package Twitch

import (
	"net/http"
	"net/url"
)

type Link struct {
	rid    string
	cid    string
	sig    string
	token  string
	client *http.Client
}

func NewTwitchLink(rid string, proxy string) (*Link, error) {
	var (
		err      error
		proxyURL *url.URL
	)
	tw := new(Link)
	tw.rid = rid
	if len(proxy) > 0 {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		tw.client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)}}
	} else {
		tw.client = &http.Client{}
	}
	err = tw.getClientID()
	if err != nil {
		return nil, err
	}
	err = tw.getSigToken()
	return tw, nil
}
