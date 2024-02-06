package main

import (
	"context"
	"flag"
	"log"

	"github.com/ddelnano/terraform-provider-xenorchestra/xoa"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

// Generate the Terraform provider documentation using `tfplugindocs`:
//
//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

func main() {

	var debugMode bool
	flag.BoolVar(&debugMode, "debuggable", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	if debugMode {
		err := plugin.Debug(context.Background(), "registry.terraform.io/vatesfr/xenorchestra",
			&plugin.ServeOpts{
				ProviderFunc: xoa.Provider,
			})
		if err != nil {
			log.Println(err.Error())
		}
	} else {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: xoa.Provider})
	}
}
