package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"reflect"
	"strings"
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

type XOClient interface {
	GetObjectsWithTags(tags []string) ([]Object, error)
	GetAllObjectsOfType(obj XoObject, response interface{}) error

	CreateVm(vmReq Vm, d time.Duration) (*Vm, error)
	GetVm(vmReq Vm) (*Vm, error)
	GetVms(vm Vm) ([]Vm, error)
	UpdateVm(vmReq Vm) (*Vm, error)
	DeleteVm(id string) error
	HaltVm(id string) error
	StartVm(id string) error

	GetCloudConfigByName(name string) ([]CloudConfig, error)
	CreateCloudConfig(name, template string) (*CloudConfig, error)
	GetCloudConfig(id string) (*CloudConfig, error)
	DeleteCloudConfig(id string) error
	GetAllCloudConfigs() ([]CloudConfig, error)

	GetHostById(id string) (host Host, err error)
	GetHostByName(nameLabel string) (hosts []Host, err error)

	GetPools(pool Pool) ([]Pool, error)
	GetPoolByName(name string) (pools []Pool, err error)

	GetSortedHosts(host Host, sortBy, sortOrder string) (hosts []Host, err error)

	CreateResourceSet(rsReq ResourceSet) (*ResourceSet, error)
	GetResourceSets() ([]ResourceSet, error)
	GetResourceSet(rsReq ResourceSet) ([]ResourceSet, error)
	GetResourceSetById(id string) (*ResourceSet, error)
	DeleteResourceSet(rsReq ResourceSet) error
	AddResourceSetSubject(rsReq ResourceSet, subject string) error
	AddResourceSetObject(rsReq ResourceSet, object string) error
	AddResourceSetLimit(rsReq ResourceSet, limit string, quantity int) error
	RemoveResourceSetSubject(rsReq ResourceSet, subject string) error
	RemoveResourceSetObject(rsReq ResourceSet, object string) error
	RemoveResourceSetLimit(rsReq ResourceSet, limit string) error

	CreateUser(user User) (*User, error)
	GetAllUsers() ([]User, error)
	GetUser(userReq User) (*User, error)
	GetCurrentUser() (*User, error)
	DeleteUser(userReq User) error

	CreateNetwork(netReq CreateNetworkRequest) (*Network, error)
	GetNetwork(netReq Network) (*Network, error)
	UpdateNetwork(netReq UpdateNetworkRequest) (*Network, error)
	GetNetworks() ([]Network, error)
	DeleteNetwork(id string) error

	GetPIF(pifReq PIF) (pifs []PIF, err error)
	GetPIFByDevice(dev string, vlan int) ([]PIF, error)

	GetStorageRepository(sr StorageRepository) ([]StorageRepository, error)
	GetStorageRepositoryById(id string) (StorageRepository, error)

	GetTemplate(template Template) ([]Template, error)

	GetAllVDIs() ([]VDI, error)
	GetVDIs(vdiReq VDI) ([]VDI, error)
	GetVDI(vdiReq VDI) (VDI, error)
	CreateVDI(vdiReq CreateVDIReq) (VDI, error)
	UpdateVDI(d Disk) error
	DeleteVDI(id string) error

	CreateAcl(acl Acl) (*Acl, error)
	GetAcl(aclReq Acl) (*Acl, error)
	DeleteAcl(acl Acl) error

	AddTag(id, tag string) error
	RemoveTag(id, tag string) error

	GetDisks(vm *Vm) ([]Disk, error)
	CreateDisk(vm Vm, d Disk) (string, error)
	DeleteDisk(vm Vm, d Disk) error
	ConnectDisk(d Disk) error
	DisconnectDisk(d Disk) error

	GetVIF(vifReq *VIF) (*VIF, error)
	GetVIFs(vm *Vm) ([]VIF, error)
	CreateVIF(vm *Vm, vif *VIF) (*VIF, error)
	DeleteVIF(vifReq *VIF) (err error)
	DisconnectVIF(vifReq *VIF) (err error)
	ConnectVIF(vifReq *VIF) (err error)

	GetCdroms(vm *Vm) ([]Disk, error)
	EjectCd(id string) error
	InsertCd(vmId, cdId string) error
}

type Client struct {
	rpc        jsonrpc2.JSONRPC2
	httpClient http.Client
	restApiURL *url.URL
}

type Config struct {
	Url                string
	Username           string
	Password           string
	InsecureSkipVerify bool
}

var dialer = gorillawebsocket.Dialer{
	ReadBufferSize:  MaxMessageSize,
	WriteBufferSize: MaxMessageSize,
}

func GetConfigFromEnv() Config {
	var wsURL string
	var username string
	var password string
	insecure := false
	if v := os.Getenv("XOA_URL"); v != "" {
		wsURL = v
	}
	if v := os.Getenv("XOA_USER"); v != "" {
		username = v
	}
	if v := os.Getenv("XOA_PASSWORD"); v != "" {
		password = v
	}
	if v := os.Getenv("XOA_INSECURE"); v != "" {
		insecure = true
	}
	return Config{
		Url:                wsURL,
		Username:           username,
		Password:           password,
		InsecureSkipVerify: insecure,
	}
}

func NewClient(config Config) (XOClient, error) {
	wsURL := config.Url
	username := config.Username
	password := config.Password

	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
	}
	dialer.TLSClientConfig = tlsConfig

	ws, _, err := dialer.Dial(fmt.Sprintf("%s/api/", wsURL), http.Header{})

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

	var token string
	err = c.Call(context.Background(), "token.create", map[string]interface{}{}, &token)
	if err != nil {
		return nil, err
	}

	jar, err := cookiejar.New(&cookiejar.Options{})
	if err != nil {
		return nil, err
	}

	restApiURL, err := convertWebsocketURLToRestApi(wsURL)
	if err != nil {
		return nil, err
	}

	jar.SetCookies(restApiURL, []*http.Cookie{
		&http.Cookie{
			Name:   "authenticationToken",
			Value:  token,
			MaxAge: 0,
		},
	})

	httpClient := http.Client{
		Jar: jar,
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
	}
	return &Client{
		rpc:        c,
		httpClient: httpClient,
		restApiURL: restApiURL,
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
	log.Printf("[TRACE] Made rpc call `%s` with params: %v and received %+v: result with error: %v\n", method, params, callRes, err)

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

type RefreshComparison interface {
	Propagated(obj interface{}) bool
}

type XoObject interface {
	Compare(obj interface{}) bool
}

func (c *Client) getObjectTypeFilter(obj XoObject) map[string]interface{} {
	xoApiType := ""
	switch t := obj.(type) {
	case Network:
		xoApiType = "network"
	case PIF:
		xoApiType = "PIF"
	case Pool:
		xoApiType = "pool"
	case Host:
		xoApiType = "host"
	case StorageRepository:
		xoApiType = "SR"
	case Vm:
		xoApiType = "VM"
	case Template:
		xoApiType = "VM-template"
	case VIF:
		xoApiType = "VIF"
	case VBD:
		xoApiType = "VBD"
	case VDI:
		xoApiType = "VDI"
	default:
		panic(fmt.Sprintf("XO client does not support type: %T", t))
	}
	return map[string]interface{}{
		"filter": map[string]string{
			"type": xoApiType,
		},
	}
}

func (c *Client) GetAllObjectsOfType(obj XoObject, response interface{}) error {
	return c.Call("xo.getAllObjects", c.getObjectTypeFilter(obj), response)
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

	log.Printf("[DEBUG] Found the following objects for type '%v' from xo.getAllObjects: %+v\n", t, objs)

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

// This function must be used after a successful JSONRPC websocket request
// is made. This is to verify that we can trust the websocket URL is well-formed
// so our simple ws/wss -> http/https translation will be done correctly
func convertWebsocketURLToRestApi(wsURL string) (*url.URL, error) {
	if !strings.HasPrefix(wsURL, "ws") {
		return nil, fmt.Errorf("expected `%s` to begin with ws in order to munge the URL to its http/https equivalent\n", wsURL)
	}
	return url.Parse(strings.Replace(wsURL, "ws", "http", 1))
}
