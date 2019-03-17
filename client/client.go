package client

import (
	"fmt"
	"io"
	"net/http"
	"net/rpc"
	"time"

	"github.com/gorilla/websocket"
	"github.com/powerman/rpc-codec/jsonrpc2"
)

type Client struct {
	rpc *rpc.Client
}

var dialer = websocket.Dialer{
	ReadBufferSize:  MaxMessageSize,
	WriteBufferSize: MaxMessageSize,
}

func NewClient(hostname, username, password string) (*Client, error) {
	ws, _, err := dialer.Dial(fmt.Sprintf("ws://%s/api/", hostname), http.Header{})

	if err != nil {
		return nil, err
	}

	codec := jsonrpc2.NewClientCodec(&rwc{c: ws})
	c := rpc.NewClientWithCodec(codec)

	params := map[string]interface{}{
		"email":    username,
		"password": password,
	}
	var reply clientResponse
	err = c.Call("session.signInWithPassword", params, &reply)
	if err != nil {
		return nil, err
	}
	return &Client{
		rpc: c,
	}, nil
}

type clientResponse struct {
	Email string `json:"email,omitempty"`
	Id    string `json:"id,omitempty"`
	// password string `json:"password"`
	// Version  string `json:"jsonrpc"`
	// // ID      *uint64          `json:"id"`
	// Result *json.RawMessage `json:"result,omitempty"`
	// Error  *jsonrpc2.Error  `json:"error,omitempty"`
}

type rwc struct {
	r io.Reader
	c *websocket.Conn
}

func (c *rwc) Write(p []byte) (int, error) {
	fmt.Println("Writing!")
	fmt.Println(string(p))
	err := c.c.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *rwc) Read(p []byte) (int, error) {
	fmt.Printf(string(p))
	for {
		if c.r == nil {
			// Advance to next message.
			var err error
			_, c.r, err = c.c.NextReader()
			if err != nil {
				return 0, err
			}
		}
		n, err := c.r.Read(p)
		if err == io.EOF {
			// At end of message.
			c.r = nil
			if n > 0 {
				return n, nil
			} else {
				// No data read, continue to next message.
				continue
			}
		}
		return n, err
	}
}

func (c *rwc) Close() error {
	return c.c.Close()
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message,omitempty"`
		Code    int    `json:"code,omitempty"`
	} `json:"error,omitempty"`
}

type CloudConfigResponse struct {
	ErrorResponse
	Result map[string]interface{} `json:"-"`
}

const (
	// Maximum message size allowed from peer.
	MaxMessageSize = 4096
	PongWait       = 60 * time.Second
)
