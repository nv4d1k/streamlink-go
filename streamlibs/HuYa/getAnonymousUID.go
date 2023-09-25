package HuYa

import (
	"errors"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"strings"
)

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
