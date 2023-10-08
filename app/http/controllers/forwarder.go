package controllers

import (
	"fmt"
	"github.com/nv4d1k/streamlink-go/app/lib/DouYin"
	"log"
	"net/url"
	"strings"

	"github.com/nv4d1k/streamlink-go/app/lib/BiliBili"

	"github.com/nv4d1k/streamlink-go/app/lib/HuYa"
	"github.com/nv4d1k/streamlink-go/app/lib/Twitch"
	"github.com/nv4d1k/streamlink-go/app/lib/forwarder"
	"github.com/nv4d1k/streamlink-go/app/lib/m3u8"

	"github.com/nv4d1k/streamlink-go/app/lib/DouYu"

	"github.com/gin-gonic/gin"
)

func Forwarder(c *gin.Context) {
	debug := false
	if d, ok := c.Get("debug"); ok {
		debug = d.(bool)
	}
	proxy := c.GetString("proxy")
	var proxyURL *url.URL
	var err error
	if proxy != "" {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			if debug {
				log.Printf("parsing proxy error: %s\n", err.Error())
			}
			c.String(400, "invalid proxy")
			return
		}
	}

	if debug {
		log.Printf("incoming request URL path: %s\n", c.Request.URL.Path)
		log.Printf("incoming request room: %s %s\n", c.Param("platform"), c.Param("room"))
		log.Printf("incoming request query: %s\n", c.Request.URL.RawQuery)
		if proxyURL != nil {
			log.Printf("incoming request proxy: %s\n", proxyURL.String())
		}
	}
	switch strings.ToLower(c.Param("platform")) {
	case "douyu":
		link, err := DouYu.NewDouyuLink(c.Param("room"), proxyURL, debug)
		if err != nil {
			if debug {
				log.Printf("create douyu link object error: %s\n", err.Error())
			}
			c.String(500, err.Error())
			return
		}
		p := forwarder.NewFLV(link, proxyURL, debug)
		err = p.Start(c.Writer)
		if err != nil {
			if debug {
				log.Printf("starting forward flv stream error: %s\n", err.Error())
			}
			c.String(500, err.Error())
			return
		}
		return
	case "huya":
		pp := c.DefaultQuery("url", "")
		prefix := fmt.Sprintf("%s://%s%s", func() string {
			if c.Request.TLS != nil {
				return "https"
			}
			return "http"
		}(), c.Request.Host, c.Request.URL.Path)
		if pp == "" {
			link, err := HuYa.NewHuyaLink(c.Param("room"), proxyURL, debug)
			if err != nil {
				if debug {
					log.Printf("create huya link object error: %s\n", err.Error())
				}
				c.String(500, err.Error())
				return
			}
			u, err := link.GetLink()
			if err != nil {
				if debug {
					log.Printf("get huya stream link error: %s\n", err.Error())
				}
				c.String(500, err.Error())
				return
			}
			switch getStreamExt(u) {
			case "m3u8":
				p := m3u8.NewM3u8(proxyURL, debug)
				err := p.WrapPlaylist(c, u, prefix)
				if err != nil {
					if debug {
						log.Printf("starting forward m3u8 list error: %s\n", err.Error())
					}
					c.String(400, err.Error())
					return
				}
			case "flv":
				p := forwarder.NewFLV(link, proxyURL, debug)
				err = p.Start(c.Writer)
				if err != nil {
					if debug {
						log.Printf("starting forward flv stream error: %s\n", err.Error())
					}
					c.String(500, err.Error())
					return
				}
			default:
				c.String(500, "unsupported format")
				return
			}
		} else {
			p := m3u8.NewM3u8(proxyURL, debug)
			err := p.Forward(c, pp, prefix)
			if err != nil {
				if debug {
					log.Printf("starting forward hls stream error: %s\n", err.Error())
				}
				c.String(400, err.Error())
				return
			}
		}
		return
	case "twitch":
		pp := c.DefaultQuery("url", "")
		prefix := fmt.Sprintf("%s://%s%s", func() string {
			if c.Request.TLS != nil {
				return "https"
			}
			return "http"
		}(), c.Request.Host, c.Request.URL.Path)
		if pp == "" {
			link, err := Twitch.NewTwitchLink(c.Param("room"), proxyURL, debug)
			if err != nil {
				if debug {
					log.Printf("create twitch link object error: %s\n", err.Error())
				}
				c.String(500, err.Error())
				return
			}
			url, err := link.GetLink()
			if err != nil {
				if debug {
					log.Printf("get twitch stream link error: %s\n", err.Error())
				}
				c.String(500, err.Error())
				return
			}
			p := m3u8.NewM3u8(proxyURL, debug)
			err = p.ForwardM3u8(c, url, prefix)
			if err != nil {
				if debug {
					log.Printf("starting forward hls list error: %s\n", err.Error())
				}
				c.String(400, err.Error())
				return
			}
		} else {
			p := m3u8.NewM3u8(proxyURL, debug)
			err := p.Forward(c, pp, prefix)
			if err != nil {
				if debug {
					log.Printf("starting forward hls stream error: %s\n", err.Error())
				}
				c.String(400, err.Error())
				return
			}
		}
		return
	case "bilibili":
		pp := c.DefaultQuery("url", "")
		prefix := fmt.Sprintf("%s://%s%s", func() string {
			if c.Request.TLS != nil {
				return "https"
			}
			return "http"
		}(), c.Request.Host, c.Request.URL.Path)
		if pp == "" {
			link, err := BiliBili.NewBiliBiliLink(c.Param("room"), proxyURL, debug)
			if err != nil {
				if debug {
					log.Printf("create bilibili link object error: %s\n", err.Error())
				}
				c.String(500, err.Error())
				return
			}
			u, err := link.GetLink()
			if err != nil {
				if debug {
					log.Printf("get bilibili stream link error: %s\n", err.Error())
				}
				c.String(500, err.Error())
				return
			}
			switch getStreamExt(u) {
			case "m3u8":
				p := m3u8.NewM3u8(proxyURL, debug)
				err := p.WrapPlaylist(c, u, prefix)
				if err != nil {
					if debug {
						log.Printf("starting forward m3u8 list error: %s\n", err.Error())
					}
					c.String(400, err.Error())
					return
				}
			case "flv":
				p := forwarder.NewFLV(link, proxyURL, debug)
				err = p.Start(c.Writer)
				if err != nil {
					if debug {
						log.Printf("starting forward flv stream error: %s\n", err.Error())
					}
					c.String(500, err.Error())
					return
				}
			default:
				c.String(500, "unsupported format")
				return
			}
		} else {
			p := m3u8.NewM3u8(proxyURL, debug)
			err := p.Forward(c, pp, prefix)
			if err != nil {
				if debug {
					log.Printf("starting forward hls stream error: %s\n", err.Error())
				}
				c.String(400, err.Error())
				return
			}
		}
		return
	case "douyin":
		pp := c.DefaultQuery("url", "")
		prefix := fmt.Sprintf("%s://%s%s", func() string {
			if c.Request.TLS != nil {
				return "https"
			}
			return "http"
		}(), c.Request.Host, c.Request.URL.Path)
		if pp == "" {
			link, err := DouYin.NewDouYinLink(c.Param("room"), proxyURL, debug)
			if err != nil {
				if debug {
					log.Printf("create douyin link object error: %s\n", err.Error())
				}
				c.String(500, err.Error())
				return
			}
			u, err := link.GetLink()
			if err != nil {
				if debug {
					log.Printf("get douyin stream link error: %s", err.Error())
				}
				c.String(500, err.Error())
				return
			}
			switch getStreamExt(u) {
			case "m3u8":
				p := m3u8.NewM3u8(proxyURL, debug)
				err := p.WrapPlaylist(c, u, prefix)
				if err != nil {
					if debug {
						log.Printf("starting forward m3u8 list error: %s\n", err.Error())
					}
					c.String(400, err.Error())
					return
				}
			case "flv":
				p := forwarder.NewFLV(link, proxyURL, debug)
				err = p.Start(c.Writer)
				if err != nil {
					if debug {
						log.Printf("starting forward flv stream error: %s\n", err.Error())
					}
					c.String(500, err.Error())
					return
				}
			default:
				c.String(500, "unsupported format")
				return
			}
		} else {
			p := m3u8.NewM3u8(proxyURL, debug)
			err := p.Forward(c, pp, prefix)
			if err != nil {
				if debug {
					log.Printf("starting forward hls stream error: %s\n", err.Error())
				}
				c.String(400, err.Error())
				return
			}
		}
		return
	default:
		c.String(400, "unsupported platform")
		return
	}
}

func getStreamExt(urlstr string) string {
	ux, err := url.Parse(urlstr)
	if err != nil {
		return "error"
	}
	upa := strings.Split(ux.Path, "/")
	fname := upa[len(upa)-1]
	fnames := strings.Split(fname, ".")
	if len(fnames) != 2 {
		return "error"
	}
	return fnames[1]
}
