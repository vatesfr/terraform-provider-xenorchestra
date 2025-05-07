package xoa

import (
	"context"
	"fmt"

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
			"run_now": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Triggers an immediate run of the backup job for specified VMs. Populate this block to trigger a run. Change the 'nonce' to re-trigger with the same settings on a subsequent apply.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"vms": {
							Type:        schema.TypeList,
							Required:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "List of VM IDs to include in this immediate run.",
						},
						"schedule": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "",
							Description: "Optional schedule override for this immediate run. If empty, the job runs 'now' without a specific schedule context from this trigger.",
						},
						"nonce": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "A unique string. Change this value to trigger a new run even if 'vms' and 'schedule' are unchanged.",
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

	settingsMap := settingsSet.List()[0].(map[string]any)

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
		if recipients, ok := v.([]any); ok {
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
	settingsMap := map[string]any{
		"retention":             settings.Retention,
		"compression_enabled":   settings.CompressionEnabled,
		"offline_backup":        settings.OfflineBackup,
		"checkpoint_snapshot":   settings.CheckpointSnapshot,
		"remote_enabled":        settings.RemoteEnabled,
		"remote_retention":      settings.RemoteRetention,
		"report_when_fail_only": settings.ReportWhenFailOnly,
		"report_recipients":     settings.ReportRecipients,
	}

	return d.Set("settings", []any{settingsMap})
}

func handleRunNow(ctx context.Context, d *schema.ResourceData, c *v2.XOClient, jobID uuid.UUID) error {
	runNowList := d.Get("run_now").([]any)
	if len(runNowList) > 0 && runNowList[0] != nil {
		runNowData := runNowList[0].(map[string]any)
		vmsToRun := expandStringList(runNowData["vms"].([]any))
		schedule := runNowData["schedule"].(string)

		if len(vmsToRun) == 0 {
			return fmt.Errorf("'run_now.vms' cannot be empty when 'run_now' block is specified")
		}

		_, err := c.Backup().RunJobForVMs(ctx, jobID, schedule, vmsToRun)
		if err != nil {
			return fmt.Errorf("failed to trigger RunJobForVMs for job ID %s: %w", jobID, err)
		}
	}
	return nil
}

func resourceBackupCreate(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	settings := getBackupSettings(d)

	backupPayload := &payloads.BackupJob{
		Name:     d.Get("name").(string),
		Mode:     payloads.BackupJobType(d.Get("mode").(string)),
		Schedule: d.Get("schedule").(string),
		Enabled:  d.Get("enabled").(bool),
		VMs:      expandStringList(d.Get("vms").([]any)),
		Settings: settings,
	}

	backup, err := c.Backup().CreateJob(ctx, backupPayload)
	if err != nil {
		return err
	}

	d.SetId(backup.ID.String())

	if _, ok := d.GetOk("run_now"); ok {
		if err := handleRunNow(ctx, d, c, backup.ID); err != nil {
			return fmt.Errorf("backup job created (ID: %s) but failed to trigger initial run_now: %w", backup.ID, err)
		}
	}

	return resourceBackupRead(d, m)
}

func resourceBackupRead(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	backup, err := c.Backup().GetJob(ctx, d.Id())
	if err != nil {
		return err
	}

	d.Set("name", backup.Name)
	d.Set("mode", backup.Mode)
	d.Set("schedule", backup.Schedule)
	d.Set("enabled", backup.Enabled)

	var vmList []string
	switch vmsTyped := backup.VMs.(type) {
	case string:
		vmList = []string{vmsTyped}
	case []any:
		vmList = make([]string, len(vmsTyped))
		for i, vmID := range vmsTyped {
			if s, ok := vmID.(string); ok {
				vmList[i] = s
			} else {
				return fmt.Errorf("unexpected type for VM ID in list: %T", vmID)
			}
		}
	case []string:
		vmList = vmsTyped
	default:
	}
	if vmList != nil {
		if err := d.Set("vms", vmList); err != nil {
			return err
		}
	}

	if err := setBackupSettings(d, backup.Settings); err != nil {
		return err
	}

	return nil
}

func resourceBackupUpdate(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	jobID, err := uuid.FromString(d.Id())
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	backupPayload := &payloads.BackupJob{
		ID:       jobID,
		Name:     d.Get("name").(string),
		Mode:     payloads.BackupJobType(d.Get("mode").(string)),
		Schedule: d.Get("schedule").(string),
		Enabled:  d.Get("enabled").(bool),
		VMs:      expandStringList(d.Get("vms").([]any)),
		Settings: getBackupSettings(d),
	}

	hasNonTriggerChanges := d.HasChange("name") ||
		d.HasChange("mode") ||
		d.HasChange("schedule") ||
		d.HasChange("enabled") ||
		d.HasChange("vms") ||
		d.HasChange("settings")

	if hasNonTriggerChanges {
		_, err = c.Backup().UpdateJob(ctx, backupPayload)
		if err != nil {
			return err
		}
	}

	if d.HasChange("run_now") {
		runNowConfig := d.Get("run_now").([]any)
		if len(runNowConfig) > 0 && runNowConfig[0] != nil {
			if err := handleRunNow(ctx, d, c, jobID); err != nil {
				return fmt.Errorf("backup job updated but failed to trigger run_now: %w", err)
			}
		}
	}

	return resourceBackupRead(d, m)
}

func resourceBackupDelete(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	return c.Backup().DeleteJob(ctx, uuid.FromStringOrNil(d.Id()))
}

func expandStringList(list []any) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}
