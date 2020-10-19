package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
