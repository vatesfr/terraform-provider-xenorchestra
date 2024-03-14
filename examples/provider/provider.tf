# Instruct terraform to download the provider on `terraform init`
terraform {
  required_providers {
    xenorchestra = {
      source = "vatesfr/xenorchestra"
    }
    xenorchestra_token_auth = {
      source = "vatesfr/xenorchestra"
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

provider "xenorchestra_token_auth" {
  # XOA_USER and XOA_PASSWORD cannot be set, nor can their arguments
  token = "<token from XO>" # or set XOA_TOKEN environment variable
}
