package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	gorillawebsocket "github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	"github.com/sourcegraph/jsonrpc2/websocket"
)

const (
	// Maximum message size allowed from peer.
	MaxMessageSize = 4096
	PongWait       = 60 * time.Second
)

type Client struct {
	rpc jsonrpc2.JSONRPC2
}

type Config struct {
	Url      string
	Username string
	Password string
}

var dialer = gorillawebsocket.Dialer{
	ReadBufferSize:  MaxMessageSize,
	WriteBufferSize: MaxMessageSize,
}

func GetConfigFromEnv() Config {
	var url string
	var username string
	var password string
	if v := os.Getenv("XOA_URL"); v != "" {
		url = v
	}
	if v := os.Getenv("XOA_USER"); v != "" {
		username = v
	}
	if v := os.Getenv("XOA_PASSWORD"); v != "" {
		password = v
	}
	return Config{
		Url:      url,
		Username: username,
		Password: password,
	}
}

func NewClient(config Config) (*Client, error) {
	url := config.Url
	username := config.Username
	password := config.Password

	ws, _, err := dialer.Dial(fmt.Sprintf("%s/api/", url), http.Header{})

	if err != nil {
		return nil, err
	}

	objStream := websocket.NewObjectStream(ws)
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	var h jsonrpc2.Handler
	h = &handler{}
	c := jsonrpc2.NewConn(ctx, objStream, h)

	reqParams := map[string]interface{}{
		"email":    username,
		"password": password,
	}
	var reply signInResponse
	err = c.Call(ctx, "session.signInWithPassword", reqParams, &reply)
	if err != nil {
		return nil, err
	}
	return &Client{
		rpc: c,
	}, nil
}

type XoObject interface {
	Compare(obj map[string]interface{}) bool
	New(obj map[string]interface{}) XoObject
}

func (c *Client) FindFromGetAllObjects(obj XoObject) (interface{}, error) {

	xoApiType := ""
	switch t := obj.(type) {
	case PIF:
		xoApiType = "PIF"
	case Template:
		xoApiType = "VM-template"
	default:
		panic(fmt.Sprintf("XO client does not support type: %v", t))
	}
	params := map[string]interface{}{
		"filter": map[string]string{
			"type": xoApiType,
		},
	}
	var objsRes struct {
		Objects map[string]interface{} `json:"-"`
	}
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	err := c.rpc.Call(ctx, "xo.getAllObjects", params, &objsRes.Objects)

	if err != nil {
		return obj, err
	}

	found := false
	for _, resObj := range objsRes.Objects {
		v, ok := resObj.(map[string]interface{})
		if !ok {
			return obj, errors.New("Could not coerce interface{} into map")
		}

		if v["type"].(string) != xoApiType {
			continue
		}

		if obj.Compare(v) {
			found = true
			obj = obj.New(v)
		}
	}
	if !found {
		return obj, NotFound{Type: xoApiType}
	}

	return obj, nil
}

type handler struct{}

func (h *handler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	// We do not react to json rpc notifications and are only acting as a client.
	// So there is no need to handle these callbacks.
}

type signInResponse struct {
	Email string `json:"email,omitempty"`
	Id    string `json:"id,omitempty"`
}
