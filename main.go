package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/vatesfr/terraform-provider-xenorchestra/provider"
)

// version is set at compile time by the build system.
// This can be set with -ldflags "-X main.version=x.y.z"
var version = "0.0.1"

//go:generate terraform-plugin-docs generate

func main() {
	providerserver.Serve(context.Background(), provider.NewProvider(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/vatesfr/xenorchestra",
	})
}
