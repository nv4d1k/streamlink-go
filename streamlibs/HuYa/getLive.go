package HuYa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/tidwall/gjson"
	"log"
	"net/url"
)

func (l *Link) getLive() (string, error) {
	var (
		err         error
		stream_info = map[string]map[string]string{
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
			stream_info["flv"][CDN_TYPE[value.Get("sCdnType").String()]] = u.String()
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
			stream_info["hls"][CDN_TYPE[value.Get("sCdnType").String()]] = u.String()
		}
		return true
	})
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err = enc.Encode(stream_info)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
