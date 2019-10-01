package main

import (
	"github.com/ddelnano/terraform-provider-xenorchestra/xoa"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {

	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: xoa.Provider,
	})
}
