package main

import (
	"flag"

	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/vatesfr/terraform-provider-xenorchestra/xoa"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {

	var debugMode bool
	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug:        debugMode,
		ProviderAddr: "registry.terraform.io/vatesfr/xenorchestra",
		ProviderFunc: xoa.Provider,
	}

	plugin.Serve(opts)
}
