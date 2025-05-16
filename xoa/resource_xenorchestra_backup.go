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
			"vms": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"schedule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cron": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The cron expression for the backup job schedule.",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Default:     false,
							Description: "Whether the backup job schedule is enabled.",
						},
						"timezone": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The timezone for the backup job schedule.",
						},
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The name for the backup job schedule.",
						},
					},
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
			"schedule_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The UUID of the schedule for this backup job.",
			},
		},
		Create: resourceBackupCreate,
		Read:   resourceBackupRead,
		Update: resourceBackupUpdate,
		Delete: resourceBackupDelete,
	}
}

// In the tf schema, we have the schedule and enabled fields which are enough to create the schedule
// those params are part of the create backup job however when it comes to update, we need to use the schedule service
// the schedule his its own object and can be updated independently however it cannot with the edit backup job.
// So the enabled + schedule from the normal level of backup for creating is top level but then if the schedule object
// has those fields, we put the priority on the schedule object.
func getSchedule(d *schema.ResourceData) payloads.Schedule {
	schedule := payloads.Schedule{}

	scheduleSet := d.Get("schedule").(*schema.Set)
	if scheduleSet.Len() == 0 {
		return schedule
	}

	scheduleMap := scheduleSet.List()[0].(map[string]any)

	if v, ok := scheduleMap["cron"]; ok {
		schedule.Cron = v.(string)
	}

	if v, ok := scheduleMap["enabled"]; ok {
		schedule.Enabled = v.(bool)
	}

	if v, ok := scheduleMap["timezone"]; ok {
		schedule.Timezone = v.(string)
	}

	if v, ok := scheduleMap["name"]; ok {
		schedule.Name = v.(string)
	}

	return schedule
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
		if recipients, okList := v.([]any); okList {
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

func resourceBackupCreate(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	defaultSettings := getBackupSettings(d)

	backupPayload := &payloads.BackupJob{
		Name: d.Get("name").(string),
		Mode: payloads.BackupJobType(d.Get("mode").(string)),
		VMs:  expandStringList(d.Get("vms").([]any)),
		Settings: map[string]payloads.BackupSettings{
			"": defaultSettings,
		},
	}

	createdJob, err := c.Backup().CreateJob(ctx, backupPayload)
	if err != nil {
		return fmt.Errorf("failed to create backup job: %w", err)
	}

	d.SetId(createdJob.ID.String())
	schedule := getSchedule(d)
	schedule.JobID = createdJob.ID

	createdSchedule, err := c.Schedule().Create(ctx, &schedule)
	if err != nil {
		return fmt.Errorf("job created but failed to create schedule: %w", err)
	}

	d.Set("schedule_id", createdSchedule.ID.String())

	return resourceBackupRead(d, m)
}

func resourceBackupRead(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	jobID := d.Id()

	backupJob, err := c.Backup().GetJob(ctx, jobID)
	if err != nil {
		if err.Error() == fmt.Sprintf("backup job not found with id: %s", jobID) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading backup job %s: %w", jobID, err)
	}

	d.Set("name", backupJob.Name)
	d.Set("mode", string(backupJob.Mode))

	var vmList []string
	switch vms := backupJob.VMs.(type) {
	case string:
		vmList = []string{vms}
	case []any:
		vmList = make([]string, len(vms))
		for i, vm := range vms {
			if vmStr, ok := vm.(string); ok {
				vmList[i] = vmStr
			}
		}
	case []string:
		vmList = vms
	case map[string]interface{}:
		vmsMap := vms
		if idVal, ok := vmsMap["id"]; ok {
			if idStr, ok := idVal.(string); ok && idStr != "" {
				vmList = []string{idStr}
			}
		}
	}
	d.Set("vms", vmList)

	if backupJob.Settings != nil {
		if settings, ok := backupJob.Settings[""]; ok {
			setBackupSettings(d, settings)
		}
	}

	scheduleID := d.Get("schedule_id").(string)

	if scheduleID != "" {
		scheduleUUID, err := uuid.FromString(scheduleID)
		if err == nil {
			schedule, err := c.Schedule().Get(ctx, scheduleUUID)
			if err == nil {
				scheduleMap := map[string]any{
					"cron":     schedule.Cron,
					"enabled":  schedule.Enabled,
					"timezone": schedule.Timezone,
				}
				if schedule.Name != "" {
					scheduleMap["name"] = schedule.Name
				}
				d.Set("schedule", []any{scheduleMap})
				return nil
			}
		}
	}

	schedules, err := c.Schedule().GetAll(ctx)
	if err == nil && len(schedules) > 0 {
		jobUUID, _ := uuid.FromString(jobID)
		for _, schedule := range schedules {
			if schedule.JobID == jobUUID {
				d.Set("schedule_id", schedule.ID.String())
				scheduleMap := map[string]any{
					"cron":     schedule.Cron,
					"enabled":  schedule.Enabled,
					"timezone": schedule.Timezone,
				}
				if schedule.Name != "" {
					scheduleMap["name"] = schedule.Name
				}
				d.Set("schedule", []any{scheduleMap})
				return nil
			}
		}
	}

	d.Set("schedule", []any{})

	return nil
}

func resourceBackupUpdate(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	jobID, err := uuid.FromString(d.Id())
	if err != nil {
		return fmt.Errorf("invalid job ID: %w", err)
	}

	backupSettings := getBackupSettings(d)

	backupPayload := &payloads.BackupJob{
		ID:   jobID,
		Name: d.Get("name").(string),
		Mode: payloads.BackupJobType(d.Get("mode").(string)),
		VMs:  expandStringList(d.Get("vms").([]any)),
		Settings: map[string]payloads.BackupSettings{
			"": backupSettings,
		},
	}

	_, err = c.Backup().UpdateJob(ctx, backupPayload)
	if err != nil {
		return fmt.Errorf("failed to update backup job: %w", err)
	}

	if d.HasChange("schedule") || d.HasChange("enabled") || d.HasChange("name") || d.HasChanges("settings.0.timezone") {
		scheduleID := d.Get("schedule_id").(string)

		schedule := getSchedule(d)
		schedule.JobID = jobID

		if scheduleID != "" {
			scheduleUUID, err := uuid.FromString(scheduleID)
			if err != nil {
				return fmt.Errorf("invalid schedule ID: %w", err)
			}

			schedule.ID = scheduleUUID

			_, err = c.Schedule().Update(ctx, scheduleUUID, &schedule)
			if err != nil {
				return fmt.Errorf("failed to update schedule: %w", err)
			}
		} else {
			createdSchedule, err := c.Schedule().Create(ctx, &schedule)
			if err != nil {
				return fmt.Errorf("failed to create schedule: %w", err)
			}

			d.Set("schedule_id", createdSchedule.ID.String())
		}
	}

	return resourceBackupRead(d, m)
}

func resourceBackupDelete(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	scheduleID := d.Get("schedule_id").(string)

	if scheduleID != "" {
		scheduleUUID, err := uuid.FromString(scheduleID)
		if err == nil {
			if err := c.Schedule().Delete(ctx, scheduleUUID); err != nil {
				return fmt.Errorf("failed to delete schedule: %w", err)
			}
		}
	}

	return c.Backup().DeleteJob(ctx, uuid.FromStringOrNil(d.Id()))
}

func expandStringList(list []any) []string {
	result := make([]string, len(list))
	for i, v := range list {
		result[i] = v.(string)
	}
	return result
}
