package xoa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/pkg/payloads"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

func dataSourceXenorchestraBackup() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to retrieve information about Xen Orchestra backup jobs.",
		Read:        dataSourceBackupRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The ID of the backup job.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the backup job.",
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The mode of the backup job.",
			},
			"schedule": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The schedule of the backup job.",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the backup job is enabled.",
			},
			"vms": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The list of VM IDs included in the backup job.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"settings": {
				Type:        schema.TypeMap,
				Computed:    true,
				Description: "The settings for the backup job.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceBackupRead(d *schema.ResourceData, m interface{}) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	if id, ok := d.GetOk("id"); ok {
		backup, err := c.Backup().GetJob(ctx, id.(string))
		if err != nil {
			return err
		}

		d.SetId(backup.ID.String())
		return setBackupData(d, backup)
	}

	if name, ok := d.GetOk("name"); ok {
		jobs, err := c.Backup().ListJobs(ctx, 0)
		if err != nil {
			return err
		}

		for _, job := range jobs {
			if job.Name == name.(string) {
				d.SetId(job.ID.String())
				return setBackupData(d, job)
			}
		}
	}

	return nil
}

func setBackupData(d *schema.ResourceData, backup any) error {
	b := backup.(*payloads.BackupJob)

	d.Set("name", b.Name)
	d.Set("mode", string(b.Mode))
	d.Set("schedule", b.Schedule)
	d.Set("enabled", b.Enabled)
	d.Set("vms", b.VMs)

	settings := map[string]interface{}{
		"compression_enabled":   b.Settings.CompressionEnabled,
		"offline_backup":        b.Settings.OfflineBackup,
		"checkpoint_snapshot":   b.Settings.CheckpointSnapshot,
		"remote_enabled":        b.Settings.RemoteEnabled,
		"remote_retention":      b.Settings.RemoteRetention,
		"report_when_fail_only": b.Settings.ReportWhenFailOnly,
	}

	if len(b.Settings.ReportRecipients) > 0 {
		recipients := make([]string, len(b.Settings.ReportRecipients))
		for i, recipient := range b.Settings.ReportRecipients {
			recipients[i] = recipient
		}
		settings["report_recipients"] = recipients
	}

	d.Set("settings", settings)

	return nil
}
