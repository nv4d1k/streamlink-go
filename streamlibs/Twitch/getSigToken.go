package Twitch

import (
	"fmt"
	"github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"strings"
)

func (l *Link) getSigToken() error {
	data := fmt.Sprintf(`{
  "operationName": "PlaybackAccessToken_Template",
  "query": "query PlaybackAccessToken_Template($login: String!, $isLive: Boolean!, $vodID: ID!, $isVod: Boolean!, $playerType: String!) {  streamPlaybackAccessToken(channelName: $login, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isLive) {    value    signature    __typename  }  videoPlaybackAccessToken(id: $vodID, params: {platform: \"web\", playerBackend: \"mediaplayer\", playerType: $playerType}) @include(if: $isVod) {    value    signature    __typename  }}",
  "variables": {
    "isLive": true,
    "login": "%s",
    "isVod": false,
    "vodID": "",
    "playerType": "site"
  }
}`, l.rid)
	req, err := http.NewRequest("POST", "https://gql.twitch.tv/gql", strings.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", fmt.Sprintf("https://www.twitch.tv/%s", l.rid))
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36")
	req.Header.Add("Client-ID", l.cid)
	resp, err := l.client.Do(req)
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
	result := gjson.ParseBytes(body)
	l.sig = result.Get("data.streamPlaybackAccessToken.signature").String()
	l.token = result.Get("data.streamPlaybackAccessToken.value").String()
	return nil
}
