package HuYa

import (
	"errors"
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"regexp"
)

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
	req.Header.Set("User-Agent", `Mozilla/5.0 (Linux; Android 5.0; SM-G900P Build/LRX21T) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/75.0.3770.100 Mobile Safari/537.36`)
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
