package DouYu

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"net/url"
	"strings"
)

func (l *Link) getPreData() (errorCode int64, err error) {
	var (
		req  *http.Request
		resp *http.Response
		body []byte
	)
	data := url.Values{}
	data.Add("rid", l.rid)
	data.Add("did", l.did)
	hash := md5.Sum([]byte(fmt.Sprintf("%s%s", l.rid, l.t13)))
	auth := hex.EncodeToString(hash[:])
	req, err = http.NewRequest("POST", fmt.Sprintf("https://playweb.douyucdn.cn/lapi/live/hlsH5Preview/%s", l.rid), strings.NewReader(data.Encode()))
	if err != nil {
		return 0, fmt.Errorf("making request error: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("rid", l.rid)
	req.Header.Add("time", l.t13)
	req.Header.Add("auth", auth)
	resp, err = l.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("post for predata error: %w", err)
	}
	defer resp.Body.Close()
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("parse data error: %w", err)
	}
	l.apiErrCode = gjson.GetBytes(body, "error").Int()
	switch gjson.GetBytes(body, "error").Int() {
	case 0, 742017:
		return gjson.GetBytes(body, "error").Int(), nil
	default:
		return gjson.GetBytes(body, "error").Int(), fmt.Errorf("%s(%d)", gjson.GetBytes(body, "msg").String(), gjson.GetBytes(body, "error").Int())
	}
}
