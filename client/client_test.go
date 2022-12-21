package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"testing"

	"github.com/sourcegraph/jsonrpc2"
)

type jsonRPCFail struct {
	err error
}

func (rpc jsonRPCFail) Call(ctx context.Context, method string, params, result interface{}, opt ...jsonrpc2.CallOption) error {
	return rpc.err
}

func (rpc jsonRPCFail) Notify(ctx context.Context, method string, params interface{}, opt ...jsonrpc2.CallOption) error {
	return nil
}

func (rpc jsonRPCFail) Close() error {
	return nil
}

func TestCall_withJsonRPC2Error(t *testing.T) {
	var jsonRpcErr string = `{"errors":[{"code":null,"reason":"type","message":"must be string, but is object","property":"@.template"}]}`
	rpcCode := 10
	msg := "invalid parameters"
	var expectedErrMsg string = fmt.Sprintf(`jsonrpc2: code %d message: %s: %s`, rpcCode, msg, jsonRpcErr)
	var data json.RawMessage = []byte(jsonRpcErr)
	c := Client{
		rpc: jsonRPCFail{
			err: &jsonrpc2.Error{
				Data:    &data,
				Code:    int64(rpcCode),
				Message: msg,
			},
		},
	}

	params := map[string]interface{}{
		"test": "test",
	}
	err := c.Call("dummy method", params, nil)

	if err == nil {
		t.Errorf("Call method should have returned non-nil error")
	}

	if err.Error() != expectedErrMsg {
		t.Errorf("Call method should surface property with invalid parameter. Received `%s` but expected `%s`", err, expectedErrMsg)
	}
}

func TestCall_withJsonRPC2ErrorWithNilData(t *testing.T) {
	rpcCode := 10
	msg := "invalid parameters"
	var expectedErrMsg string = fmt.Sprintf(`jsonrpc2: code %d message: %s`, rpcCode, msg)
	c := Client{
		rpc: jsonRPCFail{
			err: &jsonrpc2.Error{
				Data:    nil,
				Code:    int64(rpcCode),
				Message: msg,
			},
		},
	}

	params := map[string]interface{}{
		"test": "test",
	}
	err := c.Call("dummy method", params, nil)

	if err == nil {
		t.Errorf("Call method should have returned non-nil error")
	}

	if err.Error() != expectedErrMsg {
		t.Errorf("Call method should surface property with invalid parameter. Received `%s` but expected `%s`", err, expectedErrMsg)
	}
}

func TestCall_withNonJsonRPC2Error(t *testing.T) {
	expectedErr := errors.New("This is not a jsonrpc2 error")
	c := Client{
		rpc: jsonRPCFail{
			err: expectedErr,
		},
	}

	params := map[string]interface{}{
		"test": "test",
	}
	err := c.Call("dummy method", params, nil)

	if err == nil {
		t.Errorf("Call method should have returned non-nil error")
	}

	if err != expectedErr {
		t.Errorf("Call method should return an error as is if not of type `jsonrpc2.Error`. Expected: %v received: %v", expectedErr, err)
	}
}

func Test_convertWebsocketURLToRestApi(t *testing.T) {
	tests := []struct {
		inputURL    string
		expectedURL *url.URL
		err         error
	}{
		{
			// Use a URL that contains 'ws' more than once to verify
			// string replacement only modifies the prefix
			inputURL: "ws://example.com/wss",
			expectedURL: &url.URL{
				Scheme: "http",
				Host:   "example.com",
				Path:   "/wss",
			},
			err: nil,
		},
		{
			// Use a URL that contains 'ws' more than once to verify
			// string replacement only modifies the prefix
			inputURL: "wss://example.com/wss",
			expectedURL: &url.URL{
				Scheme: "https",
				Host:   "example.com",
				Path:   "/wss",
			},
			err: nil,
		},
		{
			inputURL:    "ftp://example.com",
			expectedURL: nil,
			err:         fmt.Errorf("expected `%s` to begin with ws in order to munge the URL to its http/https equivalent\n", "ftp://example.com"),
		},
	}

	for _, tt := range tests {
		url, err := convertWebsocketURLToRestApi(tt.inputURL)

		if (tt.err == nil && err != tt.err) || (tt.err != nil && tt.err.Error() != err.Error()) {
			t.Errorf("expected error `%v` to match `%v`\n", err, tt.err)
		}

		if tt.expectedURL == nil && tt.expectedURL != url {
			t.Errorf("expected url creation to return nil but instead received `%v`\n", err)
		}

		if tt.expectedURL != nil {
			urlStr := url.String()
			expectedURLStr := tt.expectedURL.String()
			if expectedURLStr != urlStr {
				t.Errorf("expected `%s` and `%s` to match\n", expectedURLStr, urlStr)
			}
		}
	}
}
