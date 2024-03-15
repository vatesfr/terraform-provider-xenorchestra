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

	"github.com/cenkalti/backoff/v3"
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
	SuspendVm(id string) error
	PauseVm(id string) error

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
	CreateBondedNetwork(netReq CreateBondedNetworkRequest) (*Network, error)
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
	RetryMode    RetryMode
	RetryMaxTime time.Duration
	rpc          jsonrpc2.JSONRPC2
	httpClient   http.Client
	restApiURL   *url.URL
}

type RetryMode int

const (
	None RetryMode = iota // specifies that no retries will be made
	// Specifies that exponential backoff will be used for certain retryable errors. When
	// a guest is booting there is the potential for a race condition if the given action
	// relies on the existance of a PV driver (unplugging / plugging a device). This open
	// allows the provider to retry these errors until the guest is initialized.
	Backoff
)

type Config struct {
	Url                string
	Username           string
	Password           string
	Token              string
	InsecureSkipVerify bool
	RetryMode          RetryMode
	RetryMaxTime       time.Duration
}

var dialer = gorillawebsocket.Dialer{
	ReadBufferSize:  MaxMessageSize,
	WriteBufferSize: MaxMessageSize,
}

var (
	retryModeMap = map[string]RetryMode{
		"none":    None,
		"backoff": Backoff,
	}
)

func GetConfigFromEnv() Config {
	var wsURL string
	var username string
	var password string
	var token string
	insecure := false
	retryMode := None
	retryMaxTime := 5 * time.Minute
	if v := os.Getenv("XOA_URL"); v != "" {
		wsURL = v
	}
	if v := os.Getenv("XOA_USER"); v != "" {
		username = v
	}
	if v := os.Getenv("XOA_PASSWORD"); v != "" {
		password = v
	}
	if v := os.Getenv("XOA_TOKEN"); v != "" {
		token = v
	}
	if v := os.Getenv("XOA_INSECURE"); v != "" {
		insecure = true
	}
	if v := os.Getenv("XOA_RETRY_MODE"); v != "" {
		retry, ok := retryModeMap[v]
		if !ok {
			fmt.Println("[ERROR] failed to set retry mode, disabling retries")
		} else {
			retryMode = retry
		}
	}
	if v := os.Getenv("XOA_RETRY_MAX_TIME"); v != "" {
		duration, err := time.ParseDuration(v)
		if err == nil {
			retryMaxTime = duration
		} else {
			fmt.Println("[ERROR] failed to set retry mode, disabling retries")
		}
	}
	return Config{
		Url:                wsURL,
		Username:           username,
		Password:           password,
		Token:              token,
		InsecureSkipVerify: insecure,
		RetryMode:          retryMode,
		RetryMaxTime:       retryMaxTime,
	}
}

func NewClient(config Config) (XOClient, error) {
	wsURL := config.Url
	username := config.Username
	password := config.Password
	token := config.Token

	if token == "" && (username == "" || password == "") {
		return nil, fmt.Errorf("One of the following environment variable(s) must be set: XOA_USER and XOA_PASSWORD or XOA_TOKEN")
	}

	useTokenAuth := false
	if token != "" {
		useTokenAuth = true
	}

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

	reqParams := map[string]interface{}{}
	if useTokenAuth {
		reqParams["token"] = token
	} else {

		reqParams["email"] = username
		reqParams["password"] = password
	}
	var reply signInResponse
	err = c.Call(context.Background(), "session.signIn", reqParams, &reply)
	if err != nil {
		return nil, err
	}

	if !useTokenAuth {
		err = c.Call(context.Background(), "token.create", map[string]interface{}{}, &token)
		if err != nil {
			return nil, err
		}
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
		RetryMode:    config.RetryMode,
		RetryMaxTime: config.RetryMaxTime,
		rpc:          c,
		httpClient:   httpClient,
		restApiURL:   restApiURL,
	}, nil
}

func (c *Client) IsRetryableError(err jsonrpc2.Error) bool {

	if c.RetryMode == None {
		return false
	}

	// Error code 11 corresponds to an error condition where a VM is missing PV drivers.
	// https://github.com/vatesfr/xen-orchestra/blob/a3a2fda157fa30af4b93d34c99bac550f7c82bbc/packages/xo-common/api-errors.js#L95

	// During the boot process, there is a race condition where the PV drivers aren't available yet and
	// making XO api calls during this time can return a VM_MISSING_PV_DRIVERS error. These errors can
	// be treated as retryable since we want to wait until the VM has finished booting and its PV driver
	// is initialized.
	if err.Code == 11 || err.Code == 14 {
		return true
	}
	return false
}

func (c *Client) Call(method string, params, result interface{}) error {
	operation := func() error {
		err := c.rpc.Call(context.Background(), method, params, result)
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
				return backoff.Permanent(err)
			}

			if c.IsRetryableError(*rpcErr) {
				return err
			}

			data := rpcErr.Data

			if data == nil {
				return backoff.Permanent(err)
			}

			return backoff.Permanent(errors.New(fmt.Sprintf("%s: %s", err, *data)))
		}
		return nil
	}

	bo := backoff.NewExponentialBackOff()
	bo.MaxElapsedTime = c.RetryMaxTime
	return backoff.Retry(operation, bo)
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
