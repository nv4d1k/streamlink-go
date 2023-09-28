package controllers

import (
	"fmt"
	"log"
	"net/url"
	"strings"

	"github.com/nv4d1k/streamlink-go/app/lib/HuYa"
	"github.com/nv4d1k/streamlink-go/app/lib/Twitch"
	"github.com/nv4d1k/streamlink-go/app/lib/forwarder"
	"github.com/nv4d1k/streamlink-go/app/lib/m3u8"

	"github.com/nv4d1k/streamlink-go/app/lib/DouYu"

	"github.com/gin-gonic/gin"
)

func Forwarder(c *gin.Context) {
	proxy := c.GetString("proxy")
	var proxyURL *url.URL
	var err error
	if proxy != "" {
		proxyURL, err = url.Parse(proxy)
		if err != nil {
			c.String(400, "invalid proxy")
			return
		}
	}
	debug := false
	if d, ok := c.Get("debug"); ok {
		debug = d.(bool)
	}
	if debug {
		log.Printf("incoming request URL path: %s\n", c.Request.URL.Path)
		log.Printf("incoming request room: %s %s\n", c.Param("platform"), c.Param("room"))
		if proxyURL != nil {
			log.Printf("incoming request proxy: %s\n", proxyURL.String())
		}
	}
	switch strings.ToLower(c.Param("platform")) {
	case "douyu":
		link, err := DouYu.NewDouyuLink(c.Param("room"), proxyURL, debug)
		if err != nil {
			c.String(500, err.Error())
			return
		}
		p := forwarder.NewFLV(link, proxyURL, debug)
		err = p.Start(c.Writer)
		if err != nil {
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
				c.String(500, err.Error())
				return
			}
			u, err := link.GetLink()
			if err != nil {
				c.String(500, err.Error())
				return
			}
			ux, err := url.Parse(u)
			if err != nil {
				c.String(400, err.Error())
				return
			}
			upa := strings.Split(ux.Path, "/")
			fname := upa[len(upa)-1]
			fnames := strings.Split(fname, ".")
			if len(fnames) != 2 {
				c.String(500, "stream link has a invalid file name")
				return
			}
			switch fnames[1] {
			case "m3u8":
				p := m3u8.NewM3u8(proxyURL, debug)
				err = p.ForwardM3u8(c, u, prefix)
				if err != nil {
					c.String(400, err.Error())
					return
				}
				c.String(200, "OK")
			case "flv":
				p := forwarder.NewFLV(link, proxyURL, debug)
				err = p.Start(c.Writer)
				if err != nil {
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
				c.String(400, err.Error())
				return
			}
		}

		/*p := forwarder.NewFLV(link, proxyURL, debug)
		err = p.Start(c.Writer)
		if err != nil {
			c.String(500, err.Error())
			return
		}*/
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
				c.String(500, err.Error())
				return
			}
			url, err := link.GetLink()
			if err != nil {
				c.String(500, err.Error())
				return
			}
			p := m3u8.NewM3u8(proxyURL, debug)
			p.ForwardM3u8(c, url, prefix)
			if err != nil {
				c.String(400, err.Error())
				return
			}
			c.String(200, "OK")
		} else {
			p := m3u8.NewM3u8(proxyURL, debug)
			err := p.Forward(c, pp, prefix)
			if err != nil {
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
