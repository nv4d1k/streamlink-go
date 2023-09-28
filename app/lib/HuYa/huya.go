package HuYa

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/nv4d1k/streamlink-go/app/lib"
	"github.com/tidwall/gjson"
)

var CDN_TYPE = map[string]string{
	"AL": "阿里", "TX": "腾讯", "HW": "华为", "HS": "火山", "WS": "网宿", "HY": "虎牙",
}

type Link struct {
	rid    string
	uid    string
	uidi   int64
	uuid   string
	debug  bool
	res    gjson.Result
	client *http.Client
}

func NewHuyaLink(rid string, proxy string, debug bool) (*Link, error) {
	var (
		err      error
		proxyURL *url.URL
	)
	hy := new(Link)
	hy.rid = rid
	hy.debug = debug
	if len(proxy) > 0 {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			return nil, err
		}
		hy.client = &http.Client{Transport: lib.NewAddHeaderTransport(&http.Transport{Proxy: http.ProxyURL(proxyURL)}, true)}
	} else {
		hy.client = &http.Client{Transport: lib.NewAddHeaderTransport(nil, false)}
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

func (l *Link) GetLink() (string, error) {
	switch l.res.Get("roomInfo.eLiveStatus").Int() {
	case 2:
		liveInfo, err := l.getLive()
		if err != nil {
			return "", err
		}
		return liveInfo, nil
	case 3:
		liveLineURL, err := base64.StdEncoding.DecodeString(l.res.Get("roomProfile.liveLineUrl").String())
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("https:%s", liveLineURL), nil
	}
	return "", errors.New("not streaming now")
}

func (l *Link) getAnonymousUID() (err error) {
	var (
		resp *http.Response
		body []byte
	)
	data := `{
        "appId": 5002,
        "byPass": 3,
        "context": "",
        "version": "2.4",
        "data": {}
    }`
	resp, err = l.client.Post("https://udblgn.huya.com/web/anonymousLogin", "application/json", strings.NewReader(data))
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalln(err.Error())
		}
	}(resp.Body)
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	if !gjson.GetBytes(body, "data.uid").Exists() {
		return errors.New("anonymous user id not found")
	}
	l.uid = gjson.GetBytes(body, "data.uid").String()
	l.uidi = gjson.GetBytes(body, "data.uid").Int()
	return nil
}

func (l *Link) getLive() (string, error) {
	var (
		stream_info = map[string][]string{
			"flv": {},
			"hls": {},
		}
	)
	l.res.Get("roomInfo.tLiveInfo.tLiveStreamInfo.vStreamInfo.value").ForEach(func(key, value gjson.Result) bool {
		if value.Get("sFlvUrl").Exists() {
			anticode, err := l.processAntiCode(value.Get("sFlvAntiCode").String(), value.Get("sStreamName").String())
			if err != nil {
				log.Println(err.Error())
				return false
			}
			u, err := url.Parse(fmt.Sprintf("%s/%s.%s?%s",
				value.Get("sFlvUrl").String(),
				value.Get("sStreamName").String(),
				value.Get("sFlvUrlSuffix").String(),
				anticode))
			if err != nil {
				log.Println(err.Error())
				return false
			}
			u.Scheme = "https"
			stream_info["flv"] = append(stream_info["flv"], u.String())
		}
		if value.Get("sHlsUrl").Exists() {
			anticode, err := l.processAntiCode(value.Get("sHlsAntiCode").String(), value.Get("sStreamName").String())
			if err != nil {
				log.Println(err.Error())
				return false
			}
			u, err := url.Parse(fmt.Sprintf("%s/%s.%s?%s",
				value.Get("sHlsUrl").String(),
				value.Get("sStreamName").String(),
				value.Get("sHlsUrlSuffix").String(),
				anticode))
			if err != nil {
				log.Println(err.Error())
				return false
			}
			u.Scheme = "https"
			stream_info["hls"] = append(stream_info["hls"], u.String())
		}
		return true
	})
	if len(stream_info["flv"]) <= 0 {
		return "", errors.New("no flv link found")
	}
	r := rand.New(rand.NewSource(time.Now().Unix()))

	return stream_info["flv"][r.Intn(len(stream_info["flv"])-1)], nil
}

func (l *Link) getRoomInfo() (err error) {
	var (
		req  *http.Request
		resp *http.Response
		body []byte
	)
	req, err = http.NewRequest("GET", fmt.Sprintf("https://m.huya.com/%s", l.rid), nil)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", lib.DEFAULT_MOBILE_USER_AGENT)
	resp, err = l.client.Do(req)
	if err != nil {
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalln(err.Error())
		}
	}(resp.Body)
	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	re := regexp.MustCompile(`<script> window.HNF_GLOBAL_INIT = (.*) </script>`)
	result := re.FindStringSubmatch(string(body))
	if len(result) < 2 {
		return errors.New("HNF_GLOBAL_INIT not found")
	}
	l.res = gjson.Parse(result[1])
	return nil
}

func (l *Link) getUUID() {
	now := time.Now().UnixMilli()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	random := int64(r.Intn(1000-0)+0) | 0
	l.uuid = strconv.FormatInt((now%10000000000*1000+random)%4294967295, 10)
}

func (l *Link) processAntiCode(anticode string, streamname string) (params string, err error) {
	q, err := url.ParseQuery(anticode)
	if err != nil {
		return "", err
	}
	q.Set("ver", "1")
	q.Set("sv", "2110211124")
	q.Set("seqid", strconv.FormatInt(l.uidi+time.Now().UnixMilli(), 10))
	q.Set("uid", l.uid)
	q.Set("uuid", l.uuid)
	ssb := md5.Sum([]byte(fmt.Sprintf("%s|%s|%s", q.Get("seqid"), q.Get("ctype"), q.Get("t"))))
	ss := hex.EncodeToString(ssb[:])
	fm_orig, err := base64.StdEncoding.DecodeString(q.Get("fm"))
	if err != nil {
		return "", err
	}
	fm_orig_str := string(fm_orig)
	fm_orig_str = strings.Replace(fm_orig_str, "$0", l.uid, -1)
	fm_orig_str = strings.Replace(fm_orig_str, "$1", streamname, -1)
	fm_orig_str = strings.Replace(fm_orig_str, "$2", ss, -1)
	fm_orig_str = strings.Replace(fm_orig_str, "$3", q.Get("wsTime"), -1)
	wss := md5.Sum([]byte(fm_orig_str))
	q.Set("wsSecret", hex.EncodeToString(wss[:]))
	q.Del("fm")
	if q.Has("txyp") {
		q.Del("txyp")
	}
	return q.Encode(), nil
}
