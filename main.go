package main

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/xoa"
	"github.com/hashicorp/terraform/plugin"
)

func main() {

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: xoa.Provider,
	})
}
