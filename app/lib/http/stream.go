package http

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime"

	"github.com/nv4d1k/streamlink-go/app/lib"
	"github.com/nv4d1k/streamlink-go/app/lib/pipe"
)

func NewStream(url string, header http.Header, proxy *url.URL, debug bool) lib.Background {

	s := &stream{
		url:    url,
		pipe:   pipe.NewPipe(),
		header: header,
		proxy:  proxy,
		debug:  debug,
	}
	runtime.SetFinalizer(s, func(s *stream) {
		s.Close()
	})
	return s
}

type stream struct {
	url    string
	header http.Header
	proxy  *url.URL
	debug  bool
	conn   net.Conn
	pipe   *pipe.Pipe
	r      io.Reader
}

func (st *stream) request(req *http.Request, depth int) (*http.Response, io.Reader, net.Conn, error) {
	if st.debug {
		log.Printf("stream request url: %s\n", req.URL.String())
	}
	if depth > 30 {
		return nil, nil, nil, fmt.Errorf("too many redirects")
	}
	ux := req.URL
	var conn net.Conn
	var err error
	switch ux.Scheme {
	case "http":
		if ux.Port() == "" {
			ux.Host = ux.Host + ":80"
		}
		conn, err = net.Dial("tcp", ux.Host)
	case "https":
		if ux.Port() == "" {
			ux.Host = ux.Host + ":443"
		}
		conn, err = tls.Dial("tcp", ux.Host, nil)
	}
	if err != nil {
		return nil, nil, nil, fmt.Errorf("dial to request host error: %w", err)
	}
	err = req.Write(conn)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("write request error: %w", err)
	}
	c := bufio.NewReader(conn)
	resp, err := http.ReadResponse(c, req)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("read response error: %w", err)
	}
	switch resp.StatusCode {
	case http.StatusOK:
		return resp, resp.Body, conn, nil
	case 301, 302:
		l := resp.Header.Get("location")
		if l == "" {
			return nil, nil, nil, fmt.Errorf("err no redirect location")
		}
		url, err := url.Parse(l)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("err url in location")
		}
		req.URL = url
		return st.request(req, depth+1)

	default:
		return nil, nil, nil, fmt.Errorf("err: %s", resp.Status)
	}
}
func (st *stream) Start() error {

	req, err := http.NewRequest(http.MethodGet, st.url, nil)
	if err != nil {
		return fmt.Errorf("make request error: %w", err)
	}
	req.Header = st.header

	_, r, conn, err := st.request(req, 0)
	if err != nil {
		return fmt.Errorf("send request error: %w", err)
	}
	st.r = r
	st.conn = conn

	go st.ReadLoop()
	return nil
}

func (st *stream) ReadLoop() {
	for {
		buf := make([]byte, 65536)
		n, err := st.r.Read(buf)
		if err != nil {
			st.pipe.CloseWithError(err)
			return
		}
		st.pipe.Write(buf[:n])
	}
}

func (st *stream) Read(b []byte) (int, error) {
	return st.pipe.Read(b)
}

func (st *stream) Close() error {
	st.conn.Close()
	return nil
}
