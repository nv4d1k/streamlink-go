package m3u8

import (
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	libm3u8 "github.com/grafov/m3u8"
	"github.com/nv4d1k/streamlink-go/app/lib"
)

func NewM3u8(proxy *url.URL, debug bool) *m3u8 {
	r := &m3u8{
		proxy: proxy,
		hc:    &http.Client{},
		debug: debug,
	}
	if proxy != nil {
		r.hc.Transport = lib.NewAddHeaderTransport(&http.Transport{Proxy: http.ProxyURL(proxy)}, false)
	} else {
		r.hc.Transport = lib.NewAddHeaderTransport(nil, false)
	}
	return r
}

type m3u8 struct {
	proxy *url.URL
	hc    *http.Client
	debug bool
}

func (m *m3u8) WrapPlaylist(ctx *gin.Context, origin, prefix string) error {
	p := libm3u8.NewMasterPlaylist()
	ux, err := m.ConvertURL(origin, origin, prefix)
	if err != nil {
		return err
	}

	p.Append(ux, nil, libm3u8.VariantParams{
		Alternatives: []*libm3u8.Alternative{
			&libm3u8.Alternative{
				Type:       "VIDEO",
				Default:    true,
				Autoselect: "YES",
			},
		},
	})
	ctx.Writer.Header().Set("content-type", "application/vnd.apple.mpegurl")
	ctx.Writer.Header().Set("cache-control", "no-cache, no-store, private")
	ctx.Writer.Header().Set("transfer-encoding", "identity")
	ctx.String(200, p.Encode().String())
	return nil
}
func (m *m3u8) ConvertURL(origin, last, prefix string) (string, error) {
	if m.debug {
		log.Printf("convering url:\norigin: %s\nlast: %s\nprefix: %s\n", origin, last, prefix)
	}
	ou, err := url.Parse(origin)
	if err != nil {
		return "", fmt.Errorf("parse item url in m3u8 error: %w", err)
	}
	lu, err := url.Parse(last)
	if err != nil {
		return "", fmt.Errorf("parse last access url error: %w", err)
	}
	u, err := url.Parse(prefix)
	if err != nil {
		return "", fmt.Errorf("parse prefix url error: %w", err)
	}
	if ou.Scheme == "" {
		ou.Scheme = lu.Scheme
	}
	if ou.Host == "" {
		ou.Host = lu.Host
	}
	if strings.Split(ou.Path, "/")[0] != "" {
		pa := strings.Split(lu.Path, "/")
		oa := pa[:len(pa)-1]
		oa = append(oa, ou.Path)
		ou.Path = strings.Join(oa, "/")
	}
	if m.debug {
		log.Printf("origin url after modified: %s\n", ou.String())
	}
	queryString := u.Query()
	if m.proxy != nil {
		queryString.Set("proxy", m.proxy.String())
	}
	oue := base64.StdEncoding.EncodeToString([]byte(ou.String()))
	if m.debug {
		log.Printf("origin url after encoded: %s", oue)
	}
	queryString.Set("url", oue)
	u.RawQuery = queryString.Encode()
	if m.debug {
		log.Printf("url after converted: %s", u.String())
	}
	return u.String(), nil
}

func (m *m3u8) ForwardM3u8(ctx *gin.Context, url, prefix string) error {
	p, lt, err := m.GetM3u8(url)
	if err != nil {
		return fmt.Errorf("get m3u8 error: %w", err)
	}
	switch lt {
	case libm3u8.MEDIA:
		mediapl := p.(*libm3u8.MediaPlaylist)
		for _, u := range mediapl.Segments {
			if u == nil {
				continue
			}
			v, err := m.ConvertURL(u.URI, url, prefix)
			if err != nil {
				return fmt.Errorf("convert url error: %w", err)
			}
			u.URI = v
		}
		mediapl.ResetCache()
	case libm3u8.MASTER:
		masterpl := p.(*libm3u8.MasterPlaylist)
		for _, v := range masterpl.Variants {
			u, err := m.ConvertURL(v.URI, url, prefix)
			if err != nil {
				return fmt.Errorf("convert url error: %w", err)
			}
			v.URI = u
		}
		masterpl.ResetCache()
	}
	ctx.Writer.Header().Set("content-type", "application/vnd.apple.mpegurl")
	ctx.Writer.Header().Set("cache-control", "no-cache, no-store, private")
	ctx.Writer.Header().Set("transfer-encoding", "identity")
	if m.debug {
		log.Printf("write modified m3u8 list content to client:\n%s\n", p.Encode().String())
	}

	ctx.String(200, p.Encode().String())
	return nil
}

func (m *m3u8) GetM3u8(url string) (libm3u8.Playlist, libm3u8.ListType, error) {
	resp, err := m.hc.Get(url)
	if err != nil {
		return nil, 0, fmt.Errorf("get m3u8 file error: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("err got: %s", resp.Status)
	}

	p, lt, err := libm3u8.DecodeFrom(resp.Body, true)
	if err != nil {
		return nil, 0, fmt.Errorf("decode m3u8 file error: %w", err)
	}
	if m.debug {
		log.Printf("origin m3u8 list content:\n%s\n", p.Encode().String())
	}
	return p, lt, err

}

func (m *m3u8) Forward(ctx *gin.Context, uu, prefix string) error {
	rawb, err := base64.StdEncoding.DecodeString(uu)
	rawu := string(rawb)
	if err != nil {
		return fmt.Errorf("decoding url error: %w", err)
	}
	if m.debug {
		log.Printf("backend url: %s\n", rawu)
	}
	ux, err := url.Parse(rawu)
	if err != nil {
		return err
	}
	ff := strings.Split(ux.Path, ".")
	if ff[len(ff)-1] == "m3u8" {
		return m.ForwardM3u8(ctx, rawu, prefix)
	} else {
		resp, err := m.hc.Get(rawu)
		if err != nil {
			return fmt.Errorf("get backend file error: %w", err)
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("err got: %s", resp.Status)
		}
		defer resp.Body.Close()
		ctx.Status(resp.StatusCode)
		ctx.Header("content-type", "binary/octet-stream")
		ctx.Header("accept-ranges", "bytes")
		ctx.Header("cache-control", "no-cache, no-store, private")
		ctx.Writer.Header().Set("transfer-encoding", "identity")
		ctx.Writer.Flush()
		_, err = io.Copy(ctx.Writer, resp.Body)
		if err != nil {
			return fmt.Errorf("copy chunks error: %w", err)
		}
	}
	return nil
}
