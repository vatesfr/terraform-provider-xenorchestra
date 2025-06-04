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
			"remotes": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"schedule": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
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
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retention": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"compression_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"offline_backup": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"checkpoint_snapshot": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"remote_enabled": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"remote_retention": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"export_retention": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"delete_first": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"report_when_fail_only": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"report_recipients": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"timezone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"backup_report_tpl": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"merge_backups_synchronously": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"concurrency": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"max_export_rate": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"offline_snapshot": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"copy_retention": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"cbt_destroy_snapshot_data": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"nbd_concurrency": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"prefer_nbd": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"retention_pool_metadata": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"retention_xo_metadata": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"timeout": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"n_retries_vm_backup_failures": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"long_term_retention": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"daily": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"retention": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									"weekly": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"retention": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									"monthly": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"retention": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
									"yearly": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"retention": {
													Type:     schema.TypeInt,
													Required: true,
												},
											},
										},
									},
								},
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

func resourceBackupCreate(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	job := &payloads.BackupJob{
		Name:     d.Get("name").(string),
		Mode:     payloads.BackupJobType(d.Get("mode").(string)),
		VMs:      expandStringList(d.Get("vms").([]any)),
		Settings: buildBackupSettings(d),
	}

	if remotesList, ok := d.Get("remotes").([]any); ok && len(remotesList) > 0 {
		stringRemotes := expandStringList(remotesList)
		if len(stringRemotes) == 1 {
			job.Remotes = stringRemotes[0]
		} else {
			job.Remotes = stringRemotes
		}
	}

	createdJob, err := c.Backup().CreateJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to create backup job: %w", err)
	}

	d.SetId(createdJob.ID.String())

	scheduleSet := d.Get("schedule").(*schema.Set)
	scheduleMap := scheduleSet.List()[0].(map[string]any)

	schedule := &payloads.Schedule{
		JobID:    createdJob.ID,
		Enabled:  scheduleMap["enabled"].(bool),
		Timezone: scheduleMap["timezone"].(string),
	}

	if cron, ok := scheduleMap["cron"].(string); ok && cron != "" {
		schedule.Cron = cron
	}
	if timezone, ok := scheduleMap["timezone"].(string); ok && timezone != "" {
		schedule.Timezone = timezone
	}
	if name, ok := scheduleMap["name"].(string); ok && name != "" {
		schedule.Name = name
	}

	createdSchedule, err := c.Schedule().Create(ctx, schedule)
	if err != nil {
		return fmt.Errorf("failed to create schedule: %w", err)
	}

	d.Set("schedule_id", createdSchedule.ID.String())

	job.ID = createdJob.ID
	job.Schedule = createdSchedule.ID
	_, err = c.Backup().UpdateJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update backup job with schedule: %w", err)
	}

	return resourceBackupRead(d, m)
}

func resourceBackupRead(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	jobID := d.Id()

	job, err := c.Backup().GetJob(ctx, jobID, payloads.RestAPIJobQueryVM)
	if err != nil {
		if err.Error() == fmt.Sprintf("backup job not found with id: %s", jobID) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading backup job %s: %w", jobID, err)
	}

	d.Set("name", job.Name)
	d.Set("mode", string(job.Mode))
	d.Set("vms", normalizeVMs(job.VMs))

	if len(job.Settings) > 0 {
		settings := parseSettingsFromAPI(job.Settings)
		if len(settings) > 0 {
			d.Set("settings", []any{settings})
		}
	}

	scheduleID := job.ScheduleID()
	if scheduleID == uuid.Nil {
		scheduleID = job.Schedule
	}

	d.Set("schedule_id", scheduleID.String())

	// Get schedule details
	schedule, err := c.Schedule().Get(ctx, scheduleID)
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

	job := &payloads.BackupJob{
		ID:       jobID,
		Name:     d.Get("name").(string),
		Mode:     payloads.BackupJobType(d.Get("mode").(string)),
		VMs:      expandStringList(d.Get("vms").([]any)),
		Settings: buildBackupSettings(d),
	}

	if remotesList, ok := d.Get("remotes").([]any); ok && len(remotesList) > 0 {
		stringRemotes := expandStringList(remotesList)
		if len(stringRemotes) == 1 {
			job.Remotes = stringRemotes[0]
		} else {
			job.Remotes = stringRemotes
		}
	}

	if scheduleID := d.Get("schedule_id").(string); scheduleID != "" {
		if scheduleUUID, err := uuid.FromString(scheduleID); err == nil {
			job.Schedule = scheduleUUID
		}
	}

	_, err = c.Backup().UpdateJob(ctx, job)
	if err != nil {
		return fmt.Errorf("failed to update backup job: %w", err)
	}

	if d.HasChange("schedule") {
		scheduleID := d.Get("schedule_id").(string)
		scheduleUUID, err := uuid.FromString(scheduleID)
		if err != nil {
			return fmt.Errorf("invalid schedule ID: %w", err)
		}

		existingSchedule, err := c.Schedule().Get(ctx, scheduleUUID)
		if err != nil {
			return fmt.Errorf("failed to get existing schedule: %w", err)
		}

		schedule := existingSchedule
		schedule.ID = scheduleUUID

		scheduleSet := d.Get("schedule").(*schema.Set)
		scheduleMap := scheduleSet.List()[0].(map[string]any)

		if enabled, ok := scheduleMap["enabled"]; ok {
			schedule.Enabled = enabled.(bool)
		}
		if cron, ok := scheduleMap["cron"].(string); ok {
			schedule.Cron = cron
		}
		if timezone, ok := scheduleMap["timezone"].(string); ok {
			schedule.Timezone = timezone
		}
		if name, ok := scheduleMap["name"].(string); ok {
			schedule.Name = name
		}

		_, err = c.Schedule().Update(ctx, scheduleUUID, schedule)
		if err != nil {
			return fmt.Errorf("failed to update schedule: %w", err)
		}
	}

	return resourceBackupRead(d, m)
}

func resourceBackupDelete(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	if scheduleID := d.Get("schedule_id").(string); scheduleID != "" {
		if scheduleUUID, err := uuid.FromString(scheduleID); err == nil {
			c.Schedule().Delete(ctx, scheduleUUID)
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

func buildBackupSettings(d *schema.ResourceData) payloads.BackupSettings {
	settings := payloads.BackupSettings{}

	if settingsList, ok := d.GetOk("settings"); ok {
		settingsData := settingsList.([]any)[0].(map[string]any)

		setIntPtr := func(key string, target **int) {
			if v, ok := settingsData[key]; ok {
				if val := v.(int); val > 0 {
					*target = &val
				}
			}
		}

		setIntPtrAllowZero := func(key string, target **int) {
			if v, ok := settingsData[key]; ok {
				val := v.(int)
				*target = &val
			}
		}

		setBoolPtr := func(key string, target **bool) {
			if v, ok := settingsData[key]; ok {
				val := v.(bool)
				*target = &val
			}
		}

		setStringPtr := func(key string, target **string) {
			if v, ok := settingsData[key]; ok {
				if val := v.(string); val != "" {
					*target = &val
				}
			}
		}

		setIntPtrAllowZero("retention", &settings.Retention)
		setIntPtrAllowZero("remote_retention", &settings.RemoteRetention)
		setIntPtrAllowZero("export_retention", &settings.ExportRetention)

		setIntPtr("concurrency", &settings.Concurrency)
		setIntPtr("max_export_rate", &settings.MaxExportRate)
		setIntPtr("copy_retention", &settings.CopyRetention)
		setIntPtr("nbd_concurrency", &settings.NbdConcurrency)
		setIntPtr("retention_pool_metadata", &settings.RetentionPoolMetadata)
		setIntPtr("retention_xo_metadata", &settings.RetentionXOMetadata)
		setIntPtr("timeout", &settings.Timeout)
		setIntPtrAllowZero("n_retries_vm_backup_failures", &settings.NRetriesVmBackupFailures)

		setBoolPtr("compression_enabled", &settings.CompressionEnabled)
		setBoolPtr("offline_backup", &settings.OfflineBackup)
		setBoolPtr("checkpoint_snapshot", &settings.CheckpointSnapshot)
		setBoolPtr("remote_enabled", &settings.RemoteEnabled)
		setBoolPtr("delete_first", &settings.DeleteFirst)
		setBoolPtr("merge_backups_synchronously", &settings.MergeBackupsSynchronously)
		setBoolPtr("offline_snapshot", &settings.OfflineSnapshot)
		setBoolPtr("cbt_destroy_snapshot_data", &settings.CbtDestroySnapshotData)
		setBoolPtr("prefer_nbd", &settings.PreferNbd)

		setStringPtr("timezone", &settings.Timezone)
		setStringPtr("backup_report_tpl", &settings.BackupReportTpl)

		if v, ok := settingsData["report_when_fail_only"]; ok {
			reportWhen := payloads.ReportWhenAlways
			if v.(bool) {
				reportWhen = payloads.ReportWhenFailOnly
			}
			settings.ReportWhen = &reportWhen
		}

		if v, ok := settingsData["report_recipients"]; ok {
			if recipientsList, ok := v.([]any); ok && len(recipientsList) > 0 {
				recipients := make([]string, len(recipientsList))
				for i, r := range recipientsList {
					recipients[i] = r.(string)
				}
				settings.ReportRecipients = recipients
			}
		}

		if v, ok := settingsData["long_term_retention"]; ok {
			if ltrList, ok := v.([]any); ok && len(ltrList) > 0 {
				ltrMap := ltrList[0].(map[string]any)
				longTermRetention := make(payloads.LongTermRetentionObject)

				for _, period := range []string{"daily", "weekly", "monthly", "yearly"} {
					if periodData, exists := ltrMap[period]; exists {
						if periodList, ok := periodData.([]any); ok && len(periodList) > 0 {
							periodMap := periodList[0].(map[string]any)
							if retention, hasRetention := periodMap["retention"]; hasRetention {
								if retentionInt, ok := retention.(int); ok && retentionInt > 0 {
									longTermRetention[payloads.LongTermRetentionDurationKey(period)] = payloads.LongTermRetentionDuration{
										Retention: retentionInt,
										Settings:  map[string]any{},
									}
								}
							}
						}
					}
				}

				if len(longTermRetention) > 0 {
					settings.LongTermRetention = longTermRetention
				}
			}
		}
	}

	return settings
}

func normalizeVMs(vms any) []string {
	switch v := vms.(type) {
	case string:
		return []string{v}
	case []string:
		return v
	case []any:
		result := make([]string, len(v))
		for i, vm := range v {
			if vmStr, ok := vm.(string); ok {
				result[i] = vmStr
			}
		}
		return result
	case map[string]any:
		if id, ok := v["id"].(string); ok {
			return []string{id}
		}
	}
	return []string{}
}

func parseSettingsFromAPI(apiSettings map[string]any) map[string]any {
	result := make(map[string]any)

	if defaultSettings, ok := apiSettings[""].(map[string]any); ok {
		setIfExists := func(apiKey, tfKey string) {
			if val, exists := defaultSettings[apiKey]; exists && val != nil {
				switch v := val.(type) {
				case float64:
					result[tfKey] = int(v)
				case int:
					result[tfKey] = v
				case bool:
					result[tfKey] = v
				case string:
					if v != "" {
						result[tfKey] = v
					}
				default:
					result[tfKey] = val
				}
			}
		}

		fieldMappings := map[string]string{
			"retention":                 "retention",
			"remoteRetention":           "remote_retention",
			"compressionEnabled":        "compression_enabled",
			"offlineBackup":             "offline_backup",
			"checkpointSnapshot":        "checkpoint_snapshot",
			"remoteEnabled":             "remote_enabled",
			"deleteFirst":               "delete_first",
			"timezone":                  "timezone",
			"backupReportTpl":           "backup_report_tpl",
			"mergeBackupsSynchronously": "merge_backups_synchronously",
			"concurrency":               "concurrency",
			"maxExportRate":             "max_export_rate",
			"offlineSnapshot":           "offline_snapshot",
			"copyRetention":             "copy_retention",
			"cbtDestroySnapshotData":    "cbt_destroy_snapshot_data",
			"nbdConcurrency":            "nbd_concurrency",
			"preferNbd":                 "prefer_nbd",
			"retentionPoolMetadata":     "retention_pool_metadata",
			"retentionXoMetadata":       "retention_xo_metadata",
			"timeout":                   "timeout",
			"nRetriesVmBackupFailures":  "n_retries_vm_backup_failures",
		}

		for apiKey, tfKey := range fieldMappings {
			setIfExists(apiKey, tfKey)
		}

		if reportWhen, exists := defaultSettings["reportWhen"]; exists && reportWhen != nil {
			result["report_when_fail_only"] = (reportWhen == "failure")
		}

		if recipients, exists := defaultSettings["reportRecipients"]; exists && recipients != nil {
			if recipientsList, ok := recipients.([]any); ok && len(recipientsList) > 0 {
				strRecipients := make([]string, len(recipientsList))
				for i, r := range recipientsList {
					if str, ok := r.(string); ok {
						strRecipients[i] = str
					}
				}
				if len(strRecipients) > 0 {
					result["report_recipients"] = strRecipients
				}
			}
		}

		if val, exists := defaultSettings["longTermRetention"]; exists && val != nil {
			if ltrMap, ok := val.(map[string]any); ok && len(ltrMap) > 0 {
				longTermRetention := make([]map[string]any, 1)
				ltrData := make(map[string]any)

				for _, period := range []string{"daily", "weekly", "monthly", "yearly"} {
					if periodData, exists := ltrMap[period]; exists && periodData != nil {
						if periodMap, ok := periodData.(map[string]any); ok {
							if retention, hasRetention := periodMap["retention"]; hasRetention && retention != nil {
								ltrData[period] = []map[string]any{{
									"retention": retention,
								}}
							}
						}
					}
				}

				if len(ltrData) > 0 {
					longTermRetention[0] = ltrData
					result["long_term_retention"] = longTermRetention
				}
			}
		}
	}

	for key, value := range apiSettings {
		if key != "" && value != nil {
			if settingsData, ok := value.(map[string]any); ok {
				// Check if this looks like a schedule ID (UUID format) and has exportRetention
				if exportRetention, exists := settingsData["exportRetention"]; exists && exportRetention != nil {
					switch v := exportRetention.(type) {
					case float64:
						result["export_retention"] = int(v)
					case int:
						result["export_retention"] = v
					}
				}
			}
		}
	}

	return result
}
