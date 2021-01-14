# Xen Orchestra Provider

The Xen Orchestra provider is used to interact with the resources supported by [Xen Orchestra](https://github.com/vatesfr/xen-orchestra).
The provider needs to be configured with the proper credentials before it can be used.

## Requirements

* Terraform 0.12+
* Xen Orchestra 5.50.1

## Using the provider

### Using terraform 0.12 and lower

1. Install a pre-compiled release in your terraform [plugins directory](https://www.terraform.io/docs/configuration-0-11/providers.html).

2. Configure the provider with the necessary credentials
```hcl
# Configure the XenServer Provider
provider "xenorchestra" {
  # Must be ws or wss
  url      = "ws://hostname-of-server" # Or set XOA_URL environment variable
  username = "<username>"              # Or set XOA_USER environment variable
  password = "<password>"              # Or set XOA_PASSWORD environment variable
}
```

3. Ensure the provider is properly installed with a `terraform init`.

### Using terraform 0.13 and higher

1. Add the configuration so terraform can install the provider from the terraform registry and configure the provider with the necessary credentials

```hcl
# Instruct terraform to download the provider on `terraform init`
terraform {
  required_providers {
    xenorchestra = {
      source = "terra-farm/xenorchestra"
      version = "~> 0.3.0"
    }
  }
}

# Configure the XenServer Provider
provider "xenorchestra" {
  # Must be ws or wss
  url      = "ws://hostname-of-server" # Or set XOA_URL environment variable
  username = "<username>"              # Or set XOA_USER environment variable
  password = "<password>"              # Or set XOA_PASSWORD environment variable

  # This is false by default and
  # will disable ssl verification if true.
  # This is useful if your deployment uses
  # a self signed certificate but should be
  # used sparingly!
  insecure = <false|true>              # Or set XOA_INSECURE environment variable to any value
}
```

2. Ensure the provider is properly installed with a `terraform init`.

### Upgrading from terraform 0.12 to 0.13

If you were previously using a local copy of the provider (not using the terraform registry) you will need to ugprade your statefile in order to use the terraform registry.

This can be fixed with a `terraform state replace-provider -- -/xenorchestra registry.terraform.io/terra-farm/xenorchestra`. If you need to do this you will likely see the following error when upgrading to terraform 0.13.

```
Error: Failed to query available provider packages

Could not retrieve the list of available versions for provider -/xenorchestra:
provider registry registry.terraform.io does not have a provider named
registry.terraform.io/-/xenorchestra

# See what providers are in use. Notice how xenorchestra is mentioned twice: once for the registry.terraform.io and once by the state.
ddelnano@ddelnano-desktop:~/code/terraform$ terraform providers

Providers required by configuration:
.
├── provider[registry.terraform.io/ddelnano/mikrotik] 0.3.2-pre
├── provider[registry.terraform.io/terra-farm/xenorchestra] 0.3.5-pre
└── provider[registry.terraform.io/hashicorp/template]

Providers required by state:

    provider[registry.terraform.io/-/mikrotik]

    provider[registry.terraform.io/-/template]

    provider[registry.terraform.io/-/xenorchestra]

# Update the statefile to use the registry based provider
ddelnano@ddelnano-desktop:~/code/terraform$ terraform state replace-provider -- -/xenorchestra registry.terraform.io/terra-farm/xenorchestra
Terraform will perform the following actions:

  ~ Updating provider:
    - registry.terraform.io/-/xenorchestra
    + registry.terraform.io/terra-farm/xenorchestra

Changing 17 resources:

```
