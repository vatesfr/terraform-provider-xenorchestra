package xoa

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/pkg/payloads"
	v2 "github.com/vatesfr/xenorchestra-go-sdk/v2"
)

func resourceBackup() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"mode": {
				Type:     schema.TypeString,
				Required: true,
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"vms": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"settings": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retention": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"compression_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"offline_backup": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"checkpoint_snapshot": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"remote_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"remote_retention": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"report_when_fail_only": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"report_recipients": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
		Create: resourceBackupCreate,
		Read:   resourceBackupRead,
		Update: resourceBackupUpdate,
		Delete: resourceBackupDelete,
	}
}

func getBackupSettings(d *schema.ResourceData) payloads.BackupSettings {
	settings := payloads.BackupSettings{}

	settingsSet := d.Get("settings").(*schema.Set)
	if settingsSet.Len() == 0 {
		return settings
	}

	settingsMap := settingsSet.List()[0].(map[string]interface{})

	if v, ok := settingsMap["retention"]; ok {
		settings.Retention = v.(int)
	}

	if v, ok := settingsMap["compression_enabled"]; ok {
		settings.CompressionEnabled = v.(bool)
	}

	if v, ok := settingsMap["offline_backup"]; ok {
		settings.OfflineBackup = v.(bool)
	}

	if v, ok := settingsMap["checkpoint_snapshot"]; ok {
		settings.CheckpointSnapshot = v.(bool)
	}

	if v, ok := settingsMap["remote_enabled"]; ok {
		settings.RemoteEnabled = v.(bool)
	}

	if v, ok := settingsMap["remote_retention"]; ok {
		settings.RemoteRetention = v.(int)
	}

	if v, ok := settingsMap["report_when_fail_only"]; ok {
		settings.ReportWhenFailOnly = v.(bool)
	}

	if v, ok := settingsMap["report_recipients"]; ok {
		if recipients, ok := v.([]interface{}); ok {
			reportRecipients := make([]string, len(recipients))
			for i, r := range recipients {
				reportRecipients[i] = r.(string)
			}
			settings.ReportRecipients = reportRecipients
		}
	}

	return settings
}

func setBackupSettings(d *schema.ResourceData, settings payloads.BackupSettings) error {
	settingsMap := map[string]interface{}{
		"retention":             settings.Retention,
		"compression_enabled":   settings.CompressionEnabled,
		"offline_backup":        settings.OfflineBackup,
		"checkpoint_snapshot":   settings.CheckpointSnapshot,
		"remote_enabled":        settings.RemoteEnabled,
		"remote_retention":      settings.RemoteRetention,
		"report_when_fail_only": settings.ReportWhenFailOnly,
		"report_recipients":     settings.ReportRecipients,
	}

	return d.Set("settings", []interface{}{settingsMap})
}

func resourceBackupCreate(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)

	settings := getBackupSettings(d)

	backup, err := c.Backup().CreateJob(context.Background(), &payloads.BackupJob{
		Name:     d.Get("name").(string),
		Mode:     payloads.BackupJobType(d.Get("mode").(string)),
		Schedule: d.Get("schedule").(string),
		Enabled:  d.Get("enabled").(bool),
		VMs:      expandStringList(d.Get("vms").([]any)),
		Settings: settings,
	})
	if err != nil {
		return err
	}

	d.SetId(backup.ID.String())
	return resourceBackupRead(d, m)
}

func resourceBackupRead(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)

	backup, err := c.Backup().GetJob(context.Background(), d.Id())
	if err != nil {
		return err
	}

	d.Set("name", backup.Name)
	d.Set("mode", backup.Mode)
	d.Set("schedule", backup.Schedule)
	d.Set("enabled", backup.Enabled)

	vms, ok := backup.VMs.([]string)
	if ok {
		d.Set("vms", vms)
	} else if vmStr, ok := backup.VMs.(string); ok {
		d.Set("vms", []string{vmStr})
	}

	setBackupSettings(d, backup.Settings)

	return nil
}

func resourceBackupUpdate(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)

	backup, err := c.Backup().GetJob(context.Background(), d.Id())
	if err != nil {
		return err
	}

	backup.Name = d.Get("name").(string)
	backup.Mode = payloads.BackupJobType(d.Get("mode").(string))
	backup.Schedule = d.Get("schedule").(string)
	backup.Enabled = d.Get("enabled").(bool)
	backup.VMs = expandStringList(d.Get("vms").([]any))

	backup.Settings = getBackupSettings(d)

	_, err = c.Backup().UpdateJob(context.Background(), backup)
	if err != nil {
		return err
	}

	return resourceBackupRead(d, m)
}

func resourceBackupDelete(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)

	return c.Backup().DeleteJob(context.Background(), uuid.FromStringOrNil(d.Id()))
}

func expandStringList(list []interface{}) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}
