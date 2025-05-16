# xenorchestra_backup Resource

Creates or manages a Xen Orchestra backup job for VMs.

## Example Usage

```terraform
resource "xenorchestra_vm" "example_vm" {
  // Ensure you have a VM resource or data source for the ID
  name_label = "tf-example-vm-for-backup"
  template   = "your-template-id" // Replace with a valid template ID from your XO
  cpus       = 1
  memory_max = 1073741824 // 1GB

  network {
    network_id = "your-network-id" // Replace with your network ID
  }

  disk {
    sr_id      = "your-sr-id" // Replace with your SR ID
    name_label = "rootdisk"
    size       = 8589934592 // 8GB
  }
  // ... other required VM parameters ...
}

resource "xenorchestra_backup" "daily_backup" {
  name     = "Daily Backup for Example VM"
  mode     = "delta"      // "full" or "delta"
  vms      = [xenorchestra_vm.example_vm.id]

  schedule {
    cron     = "0 2 * * *"  // Daily at 2 AM (cron format)
    enabled  = true
    timezone = "Europe/Paris"
    name     = "Daily Backup Schedule"
  }

  settings {
    retention               = 7
    compression_enabled     = true
    offline_backup          = false
    checkpoint_snapshot     = true // Recommended for live backups
    remote_enabled          = false
    // remote_retention     = 3 // Only if remote_enabled is true
    report_when_fail_only   = true
    // report_recipients    = ["admin@example.com"]
  }

  depends_on = [xenorchestra_vm.example_vm]
}
```

## Argument Reference

*   `name` - (Required, String) The name of the backup job.
*   `mode` - (Required, String) The backup mode. Valid options: `"full"` or `"delta"`. (Note: `metadata` and `mirror` modes are handled by different API methods and are not supported by this resource currently).
*   `vms` - (Required, List of String) List of VM UUIDs to include in this backup job.
*   `schedule` - (Optional, Block) A single block to configure the backup schedule.
  *   `cron` - (Optional, String) Cron-like schedule string for the backup job (e.g., `"0 2 * * *"` for daily at 2 AM).
  *   `enabled` - (Optional, Bool) Whether the backup job schedule is enabled. Defaults to `false`.
  *   `timezone` - (Optional, String) The timezone for the backup job schedule (e.g., `"Europe/Paris"`).
  *   `name` - (Optional, String) The name for the backup job schedule.
*   `settings` - (Optional, Block) A single block to configure detailed backup settings. All fields within are optional.
  *   `retention` - (Optional, Int) Number of local backups to keep. Defaults to `0` (which might mean server default or infinite, check XO docs).
  *   `compression_enabled` - (Optional, Bool) Whether to enable compression for the backups (e.g., native zstd). Defaults to `false`.
  *   `offline_backup` - (Optional, Bool) Whether to perform offline backups (VM is shut down or suspended). Defaults to `false`.
  *   `checkpoint_snapshot` - (Optional, Bool) Whether to create a disk checkpoint snapshot (often required for live backups). Defaults to `false`.
  *   `remote_enabled` - (Optional, Bool) Whether to enable backup to a remote. Defaults to `false`.
  *   `remote_retention` - (Optional, Int) Number of remote backups to keep (if `remote_enabled` is true). Defaults to `0`.
  *   `report_when_fail_only` - (Optional, Bool) Send email reports only on failure. Defaults to `false` (reports always or based on other criteria).
  *   `report_recipients` - (Optional, List of String) List of email addresses for backup reports.
  *   `timezone` - (Optional, String) The timezone for the backup job schedule (e.g., `"Europe/Paris"`).

## Attributes Reference

In addition to the arguments listed above, the following computed attributes are exported:

*   `id` - (String) The UUID of the backup job created in Xen Orchestra.
*   `schedule_id` - (String) The UUID of the schedule for this backup job.

## Import

Backup jobs can be imported using their UUID:

```sh
terraform import xenorchestra_backup.daily_backup job-uuid-from-xo
```