package xoa

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// Until a user resource exists this just ensures that the user test sweeper
// is run

func init() {
	resource.AddTestSweepers("xenorchestra_user", &resource.Sweeper{
		Name: "xenorchestra_user",
		F:    client.RemoveUsersWithPrefix(accTestPrefix),
	})
}
