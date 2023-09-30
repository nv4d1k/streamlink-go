package BiliBili

import (
	"errors"
	"fmt"
	"github.com/nv4d1k/streamlink-go/app/lib"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
)

type Link struct {
	rid string

	debug  bool
	client *http.Client
}

func NewBiliBiliLink(rid string, proxy *url.URL, debug bool) (bili *Link, err error) {
	bili = new(Link)
	bili.rid = rid
	bili.debug = debug
	if proxy != nil {
		bili.client = &http.Client{Transport: lib.NewAddHeaderTransport(&http.Transport{Proxy: http.ProxyURL(proxy)}, true)}
	} else {
		bili.client = &http.Client{Transport: lib.NewAddHeaderTransport(nil, true)}
	}
	err = bili.getRoomStatus()
	if err != nil {
		return nil, fmt.Errorf("get room status error: %w", err)
	}
	if bili.debug {
		log.Printf("real room id is %s\n", bili.rid)
	}
	return bili, nil
}

func (l *Link) getRoomStatus() error {
	resp, err := l.client.Get(fmt.Sprintf("https://api.live.bilibili.com/room/v1/Room/room_init?id=%s", l.rid))
	if err != nil {
		return fmt.Errorf("get room init info error: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("parse room init info body error: %w", err)
	}
	if l.debug {
		log.Printf("room init data: %s", string(body))
	}
	if gjson.ParseBytes(body).Get("code").Int() != 0 {
		return errors.New("live streaming room not exist")
	}
	l.rid = gjson.ParseBytes(body).Get("data.room_id").String()
	if gjson.ParseBytes(body).Get("data.live_status").Int() != 1 {
		return errors.New("live streaming room is offline")
	}
	return nil
}

func (l *Link) GetLink() (string, error) {
	u, err := url.Parse("https://api.live.bilibili.com/xlive/web-room/v2/index/getRoomPlayInfo")
	if err != nil {
		return "", fmt.Errorf("parsing room play info url error: %w", err)
	}
	uq := u.Query()
	uq.Set("room_id", l.rid)
	uq.Set("protocol", "1") // 0 = http_stream(flv), 1 = http_hls(m3u8)
	uq.Set("format", "1")   // 0 = flv, 1 = ts, 2 = fmp4
	uq.Set("codec", "0")    // 0 = avc, 1 = hevc
	uq.Set("qn", "0")
	uq.Set("platform", "h5")
	uq.Set("ptype", "8")
	u.RawQuery = uq.Encode()
	if l.debug {
		log.Printf("rebuilt get room play info url: %s\n", u.String())
	}
	resp, err := l.client.Get(u.String())
	if err != nil {
		return "", fmt.Errorf("get room play info error: %w", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parsing room play info error: %w", err)
	}
	if l.debug {
		log.Printf("room play info: %s\n", string(body))
	}
	streams := gjson.ParseBytes(body).Get("data.playurl_info.playurl.stream")
	if !streams.Exists() {
		return "", errors.New("live streams not found")
	}
	streams = streams.Get(`#(protocol_name=="http_hls").format.#(format_name="ts").codec.#(codec_name="avc")`)
	max_qn := streams.Get("accept_qn").Array()[0].String()
	uq.Set("qn", max_qn)
	u.RawQuery = uq.Encode()

	if l.debug {
		log.Printf("rebuilt get max qn room play info url: %s\n", u.String())
	}
	respm, err := l.client.Get(u.String())
	if err != nil {
		return "", fmt.Errorf("get max qn room play info error: %w", err)
	}
	defer respm.Body.Close()
	bodym, err := io.ReadAll(respm.Body)
	if err != nil {
		return "", fmt.Errorf("parsing max qn room play info error: %w", err)
	}
	if l.debug {
		log.Printf("max qn room play info: %s\n", string(bodym))
	}
	stream := gjson.ParseBytes(bodym).Get(fmt.Sprintf(`data.playurl_info.playurl.stream.#(protocol_name=="http_hls").format.#(format_name=="ts").codec.#(current_qn=="%s")`, max_qn))
	return fmt.Sprintf("%s%s?%s",
		stream.Get("url_info.0.host").String(),
		stream.Get("base_url").String(),
		stream.Get("url_info.0.extra").String()), nil
}
