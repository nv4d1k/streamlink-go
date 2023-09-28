package m3u8

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"
	libm3u8 "github.com/grafov/m3u8"
	"github.com/nv4d1k/streamlink-go/app/lib"
)

func NewM3u8(tls bool, proxy *url.URL) *m3u8 {
	r := &m3u8{
		tls:   tls,
		proxy: proxy,
		hc:    &http.Client{},
	}
	if proxy != nil {
		r.hc.Transport = lib.NewAddHeaderTransport(&http.Transport{Proxy: http.ProxyURL(proxy)}, false)
	} else {
		r.hc.Transport = lib.NewAddHeaderTransport(nil, false)
	}
	return r
}

type m3u8 struct {
	tls   bool
	proxy *url.URL
	hc    *http.Client
}

func (m *m3u8) ConvertURL(origin string, host, prefix string) (string, error) {
	u, err := url.Parse(origin)
	if err != nil {
		return "", err
	}
	u.Path = prefix + "/" + u.Scheme + "/" + u.Host + u.Path
	u.Host = host
	if m.tls {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}
	return u.String(), nil
}

func (m *m3u8) ForwardM3u8(ctx *gin.Context, url, prefix string) error {
	p, lt, err := m.GetM3u8(url)
	if err != nil {
		return err
	}
	switch lt {
	case libm3u8.MEDIA:
		mediapl := p.(*libm3u8.MediaPlaylist)
		for _, u := range mediapl.Segments {
			if u == nil {
				continue
			}
			v, err := m.ConvertURL(u.URI, ctx.Request.Host, prefix)
			if err != nil {
				return err
			}
			u.URI = v
		}
	case libm3u8.MASTER:
		masterpl := p.(*libm3u8.MasterPlaylist)
		for _, v := range masterpl.Variants {
			u, err := m.ConvertURL(v.URI, ctx.Request.Host, prefix)
			if err != nil {
				return err
			}
			v.URI = u
		}

	}
	ctx.Writer.Header().Set("content-type", "application/vnd.apple.mpegurl")
	ctx.Writer.Header().Set("cache-control", "no-cache, no-store, private")
	ctx.String(200, p.Encode().String())
	return nil
}

func (m *m3u8) GetM3u8(url string) (libm3u8.Playlist, libm3u8.ListType, error) {
	resp, err := m.hc.Get(url)
	if err != nil {
		return nil, 0, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("err got: %s", resp.Status)
	}

	p, lt, err := libm3u8.DecodeFrom(resp.Body, true)
	if err != nil {
		return nil, 0, err
	}
	return p, lt, err

}

func (m *m3u8) Forward(ctx *gin.Context, uu, prefix string) error {
	xx := strings.Split(uu, "/")
	if len(xx) < 4 {
		return fmt.Errorf("malformed url")
	}
	sc, u := xx[1], xx[2]
	path := ""
	for _, x := range xx[3:] {
		path += "/" + x
	}
	ux := &url.URL{}
	ux.Scheme = sc
	ux.Host = u
	ux.Path = path
	rawu := ux.String()
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
			return err
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("err got: %s", resp.Status)
		}
		defer resp.Body.Close()
		ctx.Status(resp.StatusCode)
		ctx.Header("content-type", "binary/octet-stream")
		ctx.Header("accept-ranges", "bytes")
		ctx.Header("cache-control", "no-cache, no-store, private")
		ctx.Writer.Flush()
		io.Copy(ctx.Writer, resp.Body)

	}
	return nil
}
