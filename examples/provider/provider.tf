# Instruct terraform to download the provider on `terraform init`
terraform {
  required_providers {
    xenorchestra = {
      source  = "vatesfr/xenorchestra"
      version = ">= 0.35.0, < 1.0.0"
    }
  }
}

# Configure the XenServer Provider
provider "xenorchestra" {
  # Must be ws or wss : "wss://hostname-of-server"
  url      = var.xo_url                               # Or set XOA_URL environment variable

  # The provider supports both token and username/password authentication.
  # If a token is provided, it will be used for authentication and the username and password will be ignored.
  username = local.use_token ? null : var.xo_username # Or set XOA_USER environment variable
  password = local.use_token ? null : var.xo_password # Or set XOA_PASSWORD environment variable
  token    = local.use_token ? var.xo_token : null    # Or set XOA_TOKEN environment variable

  # This is false by default and
  # will disable ssl verification if true.
  # This is useful if your deployment uses
  # a self signed certificate but should be
  # used sparingly!
  insecure = var.xo_insecure
}

locals {
  use_token = var.xo_token != null && trimspace(nonsensitive(var.xo_token)) != ""
}

variable "xo_url" {
  description = "XenOrchestra URL"
  type        = string

  /* Comment this part if you want to be asked for the URL at runtime instead of using an environment variable */
  default     = null
  nullable    = true
}

variable "xo_token" {
  description = "XenOrchestra API token"
  type        = string
  sensitive   = true

  /* Comment this part if you want to be asked for the token at runtime instead of using an environment variable */
  default     = null
  nullable    = true
}

variable "xo_username" {
  description = "XenOrchestra username"
  type        = string
  
  /* Comment this part if you want to be asked for the username at runtime instead of using an environment variable */
  default     = null
  nullable    = true
}

variable "xo_password" {
  description = "XenOrchestra password"
  type        = string
  sensitive   = true
  /* Comment this part if you want to be asked for the password at runtime instead of using an environment variable */
  default     = null
  nullable    = true
}

variable "xo_insecure" {
  description = "Whether to skip TLS verification for XenOrchestra"
  type        = bool
  default     = null  # Allow to be configured by environment variable
  nullable    = true  # Allow to be configured by environment variable 
}
