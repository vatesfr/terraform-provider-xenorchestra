# Terraform Provider for Xen Orchestra

This is a terraform provider for [Xen Orchestra](https://github.com/vatesfr/xen-orchestra).

## Docs

The terraform provider is documented on the terraform registry [page](https://registry.terraform.io/providers/vatesfr/xenorchestra/latest)

General documentation about providers in terraform can be found at [the terraform documentation](https://www.terraform.io/docs/configuration/providers.html).

## Install

If using terraform 0.13, terraform is able to install the provider automatically for you. See [this](docs/index.md) for more details.

If using terraform 0.12 or lower download a suitable binary from the [releases](https://github.com/vatesfr/terraform-provider-xenorchestra/releases) and copy it to `~/.terraform.d/plugins/terraform-provider-xenorchestra_vX.Y.Z` where `X.Y.Z` is the version.

## Debugging and Logs

The provider supports detailed logging for troubleshooting and debugging purposes.

### Enable Provider Logs

To enable debug logging, set the `TF_LOG_PROVIDER` environment variable:

```bash
export TF_LOG_PROVIDER=DEBUG
terraform plan
```

### Terraform Log Levels

You can control the level of provider logging with the `TF_LOG_PROVIDER` environment variable:

```bash
export TF_LOG_PROVIDER=DEBUG
terraform apply
```

Valid `TF_LOG_PROVIDER` levels are: `TRACE`, `DEBUG`, `INFO`, `WARN`, `ERROR`.

### Log to File

To save logs to a file for analysis:

```bash
export TF_LOG_PROVIDER=DEBUG
export TF_LOG_PATH=./terraform.log
terraform apply
```

**Note**: Only enable debug logging when troubleshooting, as it can significantly increase log verbosity and may impact performance.

## Support

You can discuss any issues you have or feature requests in [Discord](https://discord.gg/ZpNq8ez).
