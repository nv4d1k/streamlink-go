package xp2p

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"time"

	ws "github.com/gorilla/websocket"
	"github.com/nv4d1k/streamlink-go/app/lib"
	"github.com/nv4d1k/streamlink-go/app/lib/pipe"
)

func NewXP2PClient(u string, header http.Header, proxy *url.URL, debug bool) lib.Background {
	c := &client{
		url:    u,
		header: header,
		dialer: &ws.Dialer{},
		pipe:   pipe.NewPipe(),
		debug:  debug,
	}
	if proxy != nil {
		c.dialer.Proxy = http.ProxyURL(proxy)
	}
	runtime.SetFinalizer(c, func(c *client) {
		c.Close()
	})
	return c
}

type client struct {
	mu     sync.Mutex
	url    string
	header http.Header
	dialer *ws.Dialer
	conn   *ws.Conn
	stopCh chan struct{}
	pipe   *pipe.Pipe
	debug  bool
}

func (c *client) Start() error {
	ctx := context.TODO()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err := c.DialContext(ctx)
	if err != nil {
		return fmt.Errorf("dail context error: %w", err)
	}
	go c.ReadLoop()
	return nil
}

func (c *client) DialContext(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	conn, resp, err := c.dialer.DialContext(ctx, c.url, c.header)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusSwitchingProtocols {
		return fmt.Errorf("dial err: %s", resp.Status)
	}

	c.conn = conn
	return nil
}

func (c *client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return nil
	}
	err := c.conn.Close()
	c.conn = nil
	return err
}

func (c *client) ReadLoop() {
	for {
		mt, body, err := c.conn.ReadMessage()
		if err != nil {
			c.pipe.CloseWithError(err)
			return
		}

		switch mt {
		case ws.BinaryMessage:
			c.pipe.Write(body)
		case ws.TextMessage:
		case ws.CloseMessage:
			c.pipe.CloseWithError(err)
			return
		default:
			c.pipe.CloseWithError(fmt.Errorf("unknown msg type: %d", mt))
			return
		}
	}
}

func (c *client) Read(b []byte) (int, error) {
	return c.pipe.Read(b)
}
