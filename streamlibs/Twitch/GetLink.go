package Twitch

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
)

func (l *Link) GetLink() (string, error) {
	params := url.Values{}
	params.Add("allow_source", "true")
	params.Add("fast_bread", "true")
	params.Add("player_backend", "mediaplayer")
	params.Add("playlist_include_framerate", "true")
	params.Add("reassignments_supported", "true")
	params.Add("sig", l.sig)
	params.Add("supported_codecs", "vp09,avc1")
	params.Add("token", l.token)
	params.Add("cdm", "wv")
	params.Add("player_version", "1.18.0")
	stream_info := map[string]string{}
	stream_info["m3u8"] = fmt.Sprintf("https://usher.ttvnw.net/api/channel/hls/%s.m3u8?%s", l.rid, params.Encode())
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "    ")
	err := enc.Encode(stream_info)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
