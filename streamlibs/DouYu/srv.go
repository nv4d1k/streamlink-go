package DouYu

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"

	ws "github.com/gorilla/websocket"
)

type Server interface {
	Play()
	Stop()
}

type srv struct {
	laddr string
	proxy string
	debug bool
}

func NewServer(listenAddr, proxy string, debug bool) Server {
	s := &srv{
		laddr: listenAddr,
		proxy: proxy,
		debug: debug,
	}
	return s
}

func (s *srv) listen() error {
	ln, err := net.Listen("tcp", s.laddr)
	if err != nil {
		return err
	}
	fmt.Println("start listen", ln.Addr().String())
	_, _ = fmt.Printf("access in player with room id. eg. http://%s/yinzitv\n\n", ln.Addr().String())
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleDyLive)
	go func() {
		err = http.Serve(ln, mux)
		if err != nil {
			panic(err.Error())
		}
	}()
	return err
}

func (s *srv) handleDyLive(w http.ResponseWriter, req *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		w.WriteHeader(400)
		return
	}
	pps := strings.Split(req.URL.Path, "/")
	if len(pps) != 2 {
		w.WriteHeader(400)
		return
	}
	var rid string
	nn := strings.Split(pps[1], ".")
	switch len(nn) {
	case 1:
		fallthrough
	case 2:
		rid = nn[0]
	default:
		w.WriteHeader(400)
		return
	}
	l, err := NewDouyuLink(rid, s.proxy, s.debug)
	if err != nil {
		w.WriteHeader(402)
		w.Write([]byte(err.Error()))
		return
	}
	url, err := l.getLink()
	if err != nil {
		w.WriteHeader(403)
		w.Write([]byte(err.Error()))
		return
	}

	w.Header().Set("content-type", "video/x-flv")
	w.Header().Set("transfer-encoding", "identity")
	w.Header().Set("connection", "close")
	w.Header().Set("cache-control", "no-cache")

	switch {
	case strings.HasPrefix(url, "http"):
		dataChannel, exitSignal := s.forwardingRequests(url)
		s.forwardingResponses(w, dataChannel)
		exitSignal <- true
	default:

		w.WriteHeader(200)
		conn, _, err := hj.Hijack()
		if err != nil {
			if s.debug {
				fmt.Println(err.Error())
			}
			conn.Close()
			return
		}
		st := s.estabStream(url)
		st.StartRead()
		defer st.Close()
		for {
			ch, eh := st.getchans()
			select {
			case err := <-eh:
				conn.Close()
				if s.debug {
					fmt.Println(err.Error())
				}
				return
			case d := <-ch:
				_, err := conn.Write(d)
				if err != nil {
					if s.debug {
						fmt.Println(err.Error())
					}
					return
				}
			}
		}
	}
}

func (s *srv) Play() {
	err := s.listen()
	if err != nil {
		panic(err.Error())
	}
}

func (s *srv) Stop() {
}

func (s *srv) forwardingRequests(u string) (<-chan []byte, chan bool) {
	dataChannel := make(chan []byte, 10)
	exitSignal := make(chan bool)
	go func() {
		var client *http.Client
		if len(s.proxy) > 0 {
			proxyURL, err := url.Parse(s.proxy)
			if err != nil {
				log.Println(err.Error())
				return
			}
			client = &http.Client{Transport: NewAddHeaderTransport(&http.Transport{Proxy: http.ProxyURL(proxyURL)})}
		} else {
			client = &http.Client{Transport: NewAddHeaderTransport(nil)}
		}
		res, err := client.Get(u)
		if err != nil {
			return
		}
		defer res.Body.Close()

		for {
			select {
			case <-exitSignal:
				return
			default:
				{
					buf := make([]byte, 4096)
					n, err := res.Body.Read(buf)
					if err != nil || n == 0 {
						fmt.Printf("recv data error: %s\n", err.Error())
						return
					} else {
						dataChannel <- buf[:n]
					}
				}
			}
		}
	}()

	return dataChannel, exitSignal
}

func (s *srv) forwardingResponses(w http.ResponseWriter, dataChannel <-chan []byte) {
	for {
		responseData, ok := <-dataChannel
		if !ok {
			fmt.Println("read data error!")
			break
		}
		n, err := w.Write(responseData)
		if err != nil || n == 0 {
			fmt.Printf("send data error: %s\n", err.Error())
			break
		}
	}
}

func (s *srv) estabStream(url string) *stream {
	st := newStream(url, s.proxy, s.debug)
	return st
}

type stream struct {
	mu    sync.Mutex
	url   string
	proxy string
	debug bool
	conn  *ws.Conn
	cchan chan []byte
	echan chan error
}

func newStream(url string, proxy string, debug bool) *stream {

	return &stream{
		url:   url,
		debug: debug,
		cchan: make(chan []byte),
		echan: make(chan error),
		proxy: proxy,
	}
}

func (s *stream) getchans() (chan []byte, chan error) {
	return s.cchan, s.echan
}

func (s *stream) StartRead() {
	go func() {
		err := s.Read()
		if err != nil {
			s.Close()
			print(err.Error())
			return
		}
	}()
}

func (s *stream) Read() error {
	d := ws.Dialer{}
	if s.proxy != "" {
		proxyURL, err := url.Parse(s.proxy)
		if err != nil {
			return err
		}
		d.Proxy = http.ProxyURL(proxyURL)
	}
	if s.debug {
		fmt.Println("dialing:", s.url)
	}
	conn, resp, err := d.Dial(s.url, map[string][]string{
		"User-Agent":    {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36 Edg/117.0.2045.36"},
		"Cache-Control": {"no-cache"},
		"Pragma":        {"no-cache"},
	})
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusSwitchingProtocols {
		return fmt.Errorf("not 101")
	}
	s.mu.Lock()
	s.conn = conn
	s.mu.Unlock()
	defer s.Close()
	for {
		mt, body, err := s.conn.ReadMessage()
		if err != nil {
			s.echan <- err
			return err
		}
		switch mt {
		case ws.BinaryMessage:
			s.cchan <- body
		case ws.TextMessage:

		case ws.CloseMessage:
			err := errors.New("1006")
			s.echan <- err
			return err
		default:
		}
	}
}

func (s *stream) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return io.EOF
	}
	if s.debug {
		fmt.Println("closing", s.url)
	}
	return s.conn.Close()
}
