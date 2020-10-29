package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
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
	var h jsonrpc2.Handler
	h = &handler{}
	c := jsonrpc2.NewConn(context.Background(), objStream, h)

	reqParams := map[string]interface{}{
		"email":    username,
		"password": password,
	}
	var reply signInResponse
	err = c.Call(context.Background(), "session.signInWithPassword", reqParams, &reply)
	if err != nil {
		return nil, err
	}
	return &Client{
		rpc: c,
	}, nil
}

func (c *Client) Call(method string, params, result interface{}, opt ...jsonrpc2.CallOption) error {
	err := c.rpc.Call(context.Background(), method, params, result, opt...)
	var callRes interface{}
	t := reflect.TypeOf(result)
	if t == nil || t.Kind() != reflect.Ptr {
		callRes = result
	} else {
		callRes = reflect.ValueOf(result).Elem()
	}
	log.Printf("[DEBUG] Made rpc call `%s` with params: %v and received %+v: result with error: %v\n", method, params, callRes, err)

	if err != nil {
		rpcErr, ok := err.(*jsonrpc2.Error)

		if !ok {
			return err
		}

		data := rpcErr.Data

		if data == nil {
			return err
		}

		return errors.New(fmt.Sprintf("%s: %s", err, *data))
	}
	return nil
}

type XoObject interface {
	Compare(obj interface{}) bool
}

func (c *Client) GetAllObjectsOfType(obj XoObject, response interface{}) error {
	xoApiType := ""
	switch t := obj.(type) {
	case Network:
		xoApiType = "network"
	case PIF:
		xoApiType = "PIF"
	case Pool:
		xoApiType = "pool"
	case StorageRepository:
		xoApiType = "SR"
	case Vm:
		xoApiType = "VM"
	case Template:
		xoApiType = "VM-template"
	case VIF:
		xoApiType = "VIF"
	default:
		panic(fmt.Sprintf("XO client does not support type: %T", t))
	}
	params := map[string]interface{}{
		"filter": map[string]string{
			"type": xoApiType,
		},
	}
	return c.Call("xo.getAllObjects", params, response)
}

func (c *Client) FindFromGetAllObjects(obj XoObject) (interface{}, error) {
	var objsRes struct {
		Objects map[string]interface{} `json:"-"`
	}
	err := c.GetAllObjectsOfType(obj, &objsRes.Objects)
	if err != nil {
		return obj, err
	}

	found := false
	t := reflect.TypeOf(obj)
	objs := reflect.MakeSlice(reflect.SliceOf(t), 0, 0)
	for _, resObj := range objsRes.Objects {
		v, ok := resObj.(map[string]interface{})
		if !ok {
			return obj, errors.New("Could not coerce interface{} into map")
		}
		b, err := json.Marshal(v)

		if err != nil {
			return objs, err
		}
		value := reflect.New(t)
		err = json.Unmarshal(b, value.Interface())
		if err != nil {
			return objs, err
		}
		if obj.Compare(value.Elem().Interface()) {
			found = true
			objs = reflect.Append(objs, value.Elem())
		}
	}
	if !found {
		return objs, NotFound{Query: obj}
	}

	log.Printf("[DEBUG] Found the following objects from xo.getAllObjects: %+v\n", objs)

	return objs.Interface(), nil
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
