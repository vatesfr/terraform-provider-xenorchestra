package helpers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

// ProviderDataToXOClient is a helper function to convert provider data to an XO SDK client. This is used in various places across the provider to create a client from the provider configuration.
func ProviderDataToXOClient(data any, diags *diag.Diagnostics) *v2.XOClient {
	if data == nil {
		return nil
	}
	c, ok := data.(*v2.XOClient)
	if !ok {
		diags.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *v2.XOClient, got: %T. Please report this issue to the provider developers.", data),
		)
		return nil
	}
	return c
}
