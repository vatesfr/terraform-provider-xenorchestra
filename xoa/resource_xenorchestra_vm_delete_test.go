package xoa

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

// fakeDeleteVmClient records the calls made by resourceVmDeleteContext.
// Methods not overridden here panic through the nil embedded interface.
type fakeDeleteVmClient struct {
	client.XOClient

	powerState string
	haltErr    error
	calls      []string
}

func (c *fakeDeleteVmClient) GetVm(vmReq client.Vm) (*client.Vm, error) {
	c.calls = append(c.calls, "GetVm")
	return &client.Vm{Id: vmReq.Id, PowerState: c.powerState}, nil
}

func (c *fakeDeleteVmClient) HaltVm(id string) error {
	c.calls = append(c.calls, "HaltVm")
	return c.haltErr
}

func (c *fakeDeleteVmClient) DeleteVm(id string) error {
	c.calls = append(c.calls, "DeleteVm")
	return nil
}

func TestResourceVmDelete_shutdownBeforeDestroy(t *testing.T) {
	tests := []struct {
		name                  string
		shutdownBeforeDestroy bool
		powerState            string
		haltErr               error
		expectedCalls         []string
		expectedErr           string
	}{
		{
			name:          "disabled deletes the VM directly",
			powerState:    client.RunningPowerState,
			expectedCalls: []string{"DeleteVm"},
		},
		{
			name:                  "enabled halts a running VM before deleting it",
			shutdownBeforeDestroy: true,
			powerState:            client.RunningPowerState,
			expectedCalls:         []string{"GetVm", "HaltVm", "DeleteVm"},
		},
		{
			name:                  "enabled does not halt an already halted VM",
			shutdownBeforeDestroy: true,
			powerState:            client.HaltedPowerState,
			expectedCalls:         []string{"GetVm", "DeleteVm"},
		},
		{
			name:                  "enabled does not delete the VM when the clean shutdown fails",
			shutdownBeforeDestroy: true,
			powerState:            client.RunningPowerState,
			haltErr:               errors.New("PV drivers were never detected"),
			expectedCalls:         []string{"GetVm", "HaltVm"},
			expectedErr:           "the VM was not deleted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := schema.TestResourceDataRaw(t, resourceVmSchema(), map[string]interface{}{
				"shutdown_before_destroy": tt.shutdownBeforeDestroy,
			})
			d.SetId("vm-id")
			c := &fakeDeleteVmClient{powerState: tt.powerState, haltErr: tt.haltErr}

			diags := resourceVmDeleteContext(context.Background(), d, c)

			if strings.Join(c.calls, ",") != strings.Join(tt.expectedCalls, ",") {
				t.Errorf("expected calls %v, got %v", tt.expectedCalls, c.calls)
			}

			if tt.expectedErr == "" {
				if diags.HasError() {
					t.Fatalf("expected no error, got %v", diags)
				}
				if d.Id() != "" {
					t.Errorf("expected the resource id to be cleared, got %q", d.Id())
				}
				return
			}

			if !diags.HasError() || !strings.Contains(diags[0].Summary, tt.expectedErr) {
				t.Fatalf("expected an error containing %q, got %v", tt.expectedErr, diags)
			}
			if d.Id() != "vm-id" {
				t.Errorf("expected the resource id to be kept, got %q", d.Id())
			}
		})
	}
}
