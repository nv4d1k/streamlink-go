package Twitch

import (
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
)

func (l *Link) getClientID() error {
	resp, err := l.client.Get(fmt.Sprintf("https://www.twitch.tv/%s", l.rid))
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Fatalln(err.Error())
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	re := regexp.MustCompile(`clientId="(.*?)"`)
	cid := re.FindStringSubmatch(string(body))
	if len(cid) < 2 {
		return errors.New("client id not found")
	}
	l.cid = cid[1]
	return nil
}
