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
		log.Println(fmt.Sprintf("incoming request URL path: %s", c.Request.URL.Path))
		log.Println(fmt.Sprintf("incoming request room: %s %s", c.Param("platform"), c.Param("room")))
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
		link, err := HuYa.NewHuyaLink(c.Param("room"), proxy, debug)
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
	case "twitch":
		pp := c.Param("param")
		prefix := fmt.Sprintf("/%s/%s", c.Param("platform"), c.Param("room"))
		if pp == "" {
			link, err := Twitch.NewTwitchLink(c.Param("room"), proxy, debug)
			if err != nil {
				c.String(500, err.Error())
				return
			}
			url, err := link.GetLink()
			if err != nil {
				c.String(500, err.Error())
				return
			}
			p := m3u8.NewM3u8(false, proxyURL)
			p.ForwardM3u8(c, url, prefix)
			if err != nil {
				c.String(400, err.Error())
				return
			}
			c.String(200, "OK")
		} else {
			p := m3u8.NewM3u8(false, proxyURL)
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
