# xenorchestra_backup_job Resource

Creates a Xen Orchestra backup job.

## Example Usage

```terraform
resource "xenorchestra_vm" "web_server" {
  name_label       = "web-server"
  name_description = "Web server VM"
  template         = data.xenorchestra_template.template.id
  # ... other VM configuration ...
}

resource "xenorchestra_backup_job" "web_server_backup" {
  name        = "Web Server Daily Backup"
  mode        = "full"
  schedule    = "0 2 * * *"  # Daily at 2am
  enabled     = true
  vms         = [xenorchestra_vm.web_server.id]
  retention   = 7  # Keep last 7 backups
  compression = "zstd"
}
```

## Argument Reference

* `name` - (Required) The name of the backup job.
* `mode` - (Required) The backup mode. Valid options are "full" or "delta".
* `schedule` - (Required) Cron-like schedule for the backup job.
* `enabled` - (Optional) Whether the backup job is enabled. Defaults to `true`.
* `vms` - (Required) List of VM IDs to back up.
* `retention` - (Required) Number of backups to keep.
* `compression` - (Optional) Compression algorithm to use. Valid options are "native", "gzip", "zstd". Defaults to "native".
* `remotes` - (Optional) List of remote IDs to store backups.

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

* `id` - The ID of the backup job.

## Import

Backup jobs can be imported using the ID: