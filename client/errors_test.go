package client

import (
	"fmt"
	"testing"
)

func TestNotFoundErrorMessage(t *testing.T) {
	vif := VIF{
		MacAddress: "E8:61:7E:8E:F1:81",
	}
	err := NotFound{
		Query: vif,
	}

	expectedMsg := fmt.Sprintf("Could not find client.VIF with query: %+v", vif)
	msg := err.Error()

	if expectedMsg != msg {
		t.Errorf("NotFound Error() message expected to be '%s' but received '%s'", expectedMsg, msg)
	}
}
