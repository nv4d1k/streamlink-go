package forwarder

import (
	"fmt"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/nv4d1k/streamlink-go/app/lib"
	h "github.com/nv4d1k/streamlink-go/app/lib/http"
	"github.com/nv4d1k/streamlink-go/app/lib/xp2p"
)

type Link interface {
	GetLink() (string, error)
}

type flvStream struct {
	link   Link
	stopCh chan struct{}
	proxy  *url.URL
	debug  bool
}

func NewFLV(link Link, proxy *url.URL, debug bool) lib.Foreground {
	return &flvStream{
		link:   link,
		stopCh: make(chan struct{}),
		proxy:  proxy,
		debug:  debug,
	}
}

func (s *flvStream) httpHeader() http.Header {
	h := make(http.Header)
	h.Set("User-Agent", lib.DEFAULT_USER_AGENT)
	return h
}

func (s *flvStream) Start(w gin.ResponseWriter) error {
	u, err := s.link.GetLink()
	if s.debug {
		log.Println(fmt.Sprintf("backend url: %s", u))
	}
	if err != nil {
		return fmt.Errorf("get backend url error: %w", err)
	}
	ux, err := url.Parse(u)
	if err != nil {
		return fmt.Errorf("parse backend url error: %w", err)
	}
	var st lib.Background
	switch ux.Scheme {
	case "http", "https":
		st = h.NewStream(u, s.httpHeader(), s.proxy, s.debug)
	case "ws", "wss":
		st = xp2p.NewXP2PClient(u, s.httpHeader(), s.proxy, s.debug)
	default:
		return fmt.Errorf("unknown protocol: %s", ux.Scheme)
	}
	err = st.Start()
	if err != nil {
		return err
	}
	w.Header().Set("content-type", "video/x-flv")
	w.Header().Set("transfer-encoding", "identity")
	w.Header().Set("connection", "close")
	w.Header().Set("cache-control", "no-cache")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "*")
	w.Header().Set("Access-Control-Allow-Methods", "*")
	w.WriteHeader(200)
	w.Flush()

	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		conn.Close()
		return err
	}
	go func() {
		defer st.Close()
		defer conn.Close()
		for {
			buf := make([]byte, 65536)
			n, err := st.Read(buf)
			if err != nil {
				return
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				return
			}
		}
	}()
	return nil
}
