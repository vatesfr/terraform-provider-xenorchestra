package xoa

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

// Until a user resource exists this just ensures that the user test sweeper
// is run

func init() {
	resource.AddTestSweepers("xenorchestra_user", &resource.Sweeper{
		Name: "xenorchestra_user",
		F:    client.RemoveUsersWithPrefix(accTestPrefix),
	})
}
