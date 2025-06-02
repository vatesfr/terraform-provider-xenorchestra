package xoa

import (
	"context"
	"fmt"
	"time"

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
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"schedule_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the schedule these settings apply to. If omitted, settings are defaults for the job.",
						},
						"remote_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "The ID of the remote these settings apply to. If omitted, settings are defaults for the job.",
						},
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
						"timezone": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"export_retention": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"delete_first": {
							Type:     schema.TypeBool,
							Optional: true,
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

func resourceBackupCreate(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	var remotesPayload any
	if remotesList, ok := d.Get("remotes").([]any); ok && len(remotesList) > 0 {
		stringRemotes := expandStringList(remotesList)
		if len(stringRemotes) == 1 {
			remotesPayload = stringRemotes[0]
		} else if len(stringRemotes) > 1 {
			remotesPayload = stringRemotes
		}
	}

	jobSettings := getBackupSettings(d, c)

	jobCreationPayload := &payloads.BackupJob{
		Name:     d.Get("name").(string),
		Mode:     payloads.BackupJobType(d.Get("mode").(string)),
		VMs:      expandStringList(d.Get("vms").([]any)),
		Remotes:  remotesPayload,
		Settings: convertMapToBackupSettings(jobSettings),
	}

	scheduleSet := d.Get("schedule").(*schema.Set)
	if scheduleSet.Len() > 0 {
		tfSchedule := getSchedule(d)

		createdJobResp, err := c.Backup().CreateJob(ctx, jobCreationPayload)
		if err != nil {
			return fmt.Errorf("failed to create backup job: %w", err)
		}
		d.SetId(createdJobResp.ID.String())

		sdkSchedulePayload := payloads.Schedule{
			Cron:     tfSchedule.Cron,
			Enabled:  tfSchedule.Enabled,
			Name:     tfSchedule.Name,
			Timezone: tfSchedule.Timezone,
			JobID:    createdJobResp.ID,
		}

		createdSchedule, errSched := c.Schedule().Create(ctx, &sdkSchedulePayload)
		if errSched != nil {
			return fmt.Errorf("job created (%s) but failed to create schedule: %w", createdJobResp.ID.String(), errSched)
		}
		if err := d.Set("schedule_id", createdSchedule.ID.String()); err != nil {
			return fmt.Errorf("error setting schedule_id: %w", err)
		}

		fullSettings := getBackupSettings(d, c)

		createdJob := &payloads.BackupJob{
			ID:       createdJobResp.ID,
			Name:     createdJobResp.Name,
			Mode:     createdJobResp.Mode,
			VMs:      createdJobResp.VMs,
			Remotes:  createdJobResp.Remotes,
			Schedule: createdSchedule.ID,
			Settings: convertMapToBackupSettings(fullSettings),
		}

		if settingsListRaw, ok := d.GetOk("settings"); ok {
			settingsList := settingsListRaw.([]any)
			for _, settingsData := range settingsList {
				settingsSchemaMap := settingsData.(map[string]any)
				if exportRet, hasExportRetention := settingsSchemaMap["export_retention"]; hasExportRetention {
					if exportRetInt, ok := exportRet.(int); ok {
						createdJob.Settings.ExportRetention = &exportRetInt
					}
				}
			}
		}

		_, err = c.Backup().UpdateJob(ctx, createdJob)
		if err != nil {
			return fmt.Errorf("failed to update backup job: %w", err)
		}
	} else {
		createdJobResp, err := c.Backup().CreateJob(ctx, jobCreationPayload)
		if err != nil {
			return fmt.Errorf("failed to create backup job: %w", err)
		}
		d.SetId(createdJobResp.ID.String())
		if err := d.Set("schedule_id", ""); err != nil {
			return fmt.Errorf("error setting schedule_id: %w", err)
		}
	}

	if err := d.Set("name", jobCreationPayload.Name); err != nil {
		return fmt.Errorf("error setting name: %w", err)
	}
	if err := d.Set("mode", string(jobCreationPayload.Mode)); err != nil {
		return fmt.Errorf("error setting mode: %w", err)
	}
	if err := d.Set("vms", jobCreationPayload.VMs); err != nil {
		return fmt.Errorf("error setting vms: %w", err)
	}

	return resourceBackupRead(d, m)
}

func resourceBackupRead(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	jobID := d.Id()

	var backupJobResp *payloads.BackupJobResponse
	var err error

	for i := 0; i < 3; i++ {
		backupJobResp, err = c.Backup().GetJob(ctx, jobID, payloads.RestAPIJobQueryVM)
		if err == nil {
			break
		}

		if i < 2 {
			time.Sleep(2 * time.Second)
		}
	}

	if err != nil {
		if err.Error() == fmt.Sprintf("backup job not found with id: %s", jobID) {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error reading backup job %s: %w", jobID, err)
	}

	if backupJobResp.Schedule != uuid.Nil {
		if err := d.Set("schedule_id", backupJobResp.Schedule.String()); err != nil {
			return fmt.Errorf("error setting schedule_id from backup job: %w", err)
		}
	}

	if err := d.Set("name", backupJobResp.Name); err != nil {
		return fmt.Errorf("error setting name: %w", err)
	}
	if err := d.Set("mode", string(backupJobResp.Mode)); err != nil {
		return fmt.Errorf("error setting mode: %w", err)
	}

	var vmList []string
	switch vms := backupJobResp.VMs.(type) {
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
	if err := d.Set("vms", vmList); err != nil {
		return fmt.Errorf("error setting vms: %w", err)
	}

	parsedSettings := parseBackupJobSettings(backupJobResp.Settings)

	var tfSettingsList []map[string]any
	if len(parsedSettings) > 0 {
		tfSettingsList = append(tfSettingsList, parsedSettings)
	}

	if err := d.Set("settings", tfSettingsList); err != nil {
		return fmt.Errorf("error setting backup job settings: %w", err)
	}

	scheduleID := d.Get("schedule_id").(string)
	scheduleFound := false

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
				if err := d.Set("schedule", []any{scheduleMap}); err != nil {
					return fmt.Errorf("error setting schedule: %w", err)
				}
				scheduleFound = true
			}
		}
	}

	if !scheduleFound {
		schedules, err := c.Schedule().GetAll(ctx)
		if err == nil && len(schedules) > 0 {
			jobUUID, _ := uuid.FromString(jobID)
			for _, schedule := range schedules {
				if schedule.JobID == jobUUID {
					if err := d.Set("schedule_id", schedule.ID.String()); err != nil {
						return fmt.Errorf("error setting schedule_id from search: %w", err)
					}
					scheduleMap := map[string]any{
						"cron":     schedule.Cron,
						"enabled":  schedule.Enabled,
						"timezone": schedule.Timezone,
					}
					if schedule.Name != "" {
						scheduleMap["name"] = schedule.Name
					}
					if err := d.Set("schedule", []any{scheduleMap}); err != nil {
						return fmt.Errorf("error setting schedule from search: %w", err)
					}
					scheduleFound = true
					break
				}
			}
		}
	}

	if !scheduleFound {
		if err := d.Set("schedule_id", ""); err != nil {
			return fmt.Errorf("error clearing schedule_id: %w", err)
		}
		if err := d.Set("schedule", []any{}); err != nil {
			return fmt.Errorf("error clearing schedule: %w", err)
		}
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

	if d.HasChange("name") || d.HasChange("mode") || d.HasChange("vms") || d.HasChange("remotes") || d.HasChange("settings") {

		var remotesPayload any
		if remotesList, ok := d.Get("remotes").([]any); ok && len(remotesList) > 0 {
			stringRemotes := expandStringList(remotesList)
			if len(stringRemotes) == 1 {
				remotesPayload = stringRemotes[0]
			} else if len(stringRemotes) > 1 {
				remotesPayload = stringRemotes
			}
		}

		backupPayload := &payloads.BackupJob{
			ID:       jobID,
			Name:     d.Get("name").(string),
			Mode:     payloads.BackupJobType(d.Get("mode").(string)),
			VMs:      expandStringList(d.Get("vms").([]any)),
			Schedule: uuid.FromStringOrNil(d.Get("schedule_id").(string)),
			Remotes:  remotesPayload,
		}

		scheduleID := d.Get("schedule_id").(string)
		if scheduleID != "" {
			if scheduleUUID, err := uuid.FromString(scheduleID); err == nil {
				backupPayload.Schedule = scheduleUUID
			}
		}

		if d.HasChange("settings") {
			terraformSettings := getBackupSettings(d, c)
			backupPayload.Settings = convertMapToBackupSettings(terraformSettings)

			if settingsListRaw, ok := d.GetOk("settings"); ok {
				settingsList := settingsListRaw.([]any)
				for _, settingsData := range settingsList {
					settingsSchemaMap := settingsData.(map[string]any)
					if exportRet, hasExportRetention := settingsSchemaMap["export_retention"]; hasExportRetention {
						if exportRetInt, ok := exportRet.(int); ok {
							backupPayload.Settings.ExportRetention = &exportRetInt
						}
					}
				}
			}
		}

		_, err = c.Backup().UpdateJob(ctx, backupPayload)
		if err != nil {
			return fmt.Errorf("failed to update backup job: %w", err)
		}
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

func getBackupSettings(d *schema.ResourceData, c *v2.XOClient) map[string]any {
	settingsMap := make(map[string]any)
	defaultSettings := make(map[string]any)

	var currentSettingsMap map[string]any
	if d.Id() != "" {
		ctx := context.Background()
		job, err := c.Backup().GetJob(ctx, d.Id(), payloads.RestAPIJobQueryVM)
		if err == nil && job.Settings != nil {
			currentSettingsMap = job.Settings
			if currentDefaults, ok := job.Settings[""].(map[string]any); ok {
				defaultSettings = currentDefaults
			}
		}
	}

	settingsListRaw, ok := d.GetOk("settings")
	if !ok {
		if currentSettingsMap != nil {
			return currentSettingsMap
		}
		settingsMap[""] = defaultSettings
		return settingsMap
	}
	settingsList := settingsListRaw.([]any)

	for _, settingsData := range settingsList {
		settingsSchemaMap := settingsData.(map[string]any)

		settingsKey := ""
		if scheduleID, ok := settingsSchemaMap["schedule_id"].(string); ok && scheduleID != "" {
			settingsKey = scheduleID
		} else if remoteID, ok := settingsSchemaMap["remote_id"].(string); ok && remoteID != "" {
			settingsKey = remoteID
		}

		var targetSettings map[string]any
		if currentSettingsMap != nil && currentSettingsMap[settingsKey] != nil {
			if existing, ok := currentSettingsMap[settingsKey].(map[string]any); ok {
				targetSettings = existing
			} else {
				targetSettings = make(map[string]any)
			}
		} else {
			targetSettings = make(map[string]any)
			if settingsKey == "" {
				targetSettings = defaultSettings
			}
		}

		hasChanged := func(tfKey string, apiKey string, newValue any) bool {
			if len(targetSettings) == 0 {
				return true
			}

			currentValue, exists := targetSettings[apiKey]
			if !exists {
				return true
			}

			settingsIndex := 0
			if tfValue, ok := d.GetOk(fmt.Sprintf("settings.%d.%s", settingsIndex, tfKey)); ok {
				return !equalValues(tfValue, currentValue)
			}

			return false
		}

		if v, ok := settingsSchemaMap["retention"]; ok {
			if valInt, okInt := v.(int); okInt && hasChanged("retention", "retention", valInt) {
				targetSettings["retention"] = valInt
			}
		}
		if v, ok := settingsSchemaMap["compression_enabled"]; ok {
			if valBool, okBool := v.(bool); okBool && hasChanged("compression_enabled", "compressionEnabled", valBool) {
				targetSettings["compressionEnabled"] = valBool
			}
		}
		if v, ok := settingsSchemaMap["offline_backup"]; ok {
			if valBool, okBool := v.(bool); okBool && hasChanged("offline_backup", "offlineBackup", valBool) {
				targetSettings["offlineBackup"] = valBool
			}
		}
		if v, ok := settingsSchemaMap["checkpoint_snapshot"]; ok {
			if valBool, okBool := v.(bool); okBool && hasChanged("checkpoint_snapshot", "checkpointSnapshot", valBool) {
				targetSettings["checkpointSnapshot"] = valBool
			}
		}
		if v, ok := settingsSchemaMap["remote_enabled"]; ok {
			if valBool, okBool := v.(bool); okBool && hasChanged("remote_enabled", "remoteEnabled", valBool) {
				targetSettings["remoteEnabled"] = valBool
			}
		}
		if v, ok := settingsSchemaMap["remote_retention"]; ok {
			if valInt, okInt := v.(int); okInt && hasChanged("remote_retention", "remoteRetention", valInt) {
				targetSettings["remoteRetention"] = valInt
			}
		}
		if v, ok := settingsSchemaMap["report_when_fail_only"]; ok {
			if valBool, okBool := v.(bool); okBool {
				currentReportWhen := ""
				if len(targetSettings) > 0 {
					if reportWhen, exists := targetSettings["reportWhen"]; exists {
						currentReportWhen = reportWhen.(string)
					}
				}

				expectedReportWhen := "always"
				if valBool {
					expectedReportWhen = "failure"
				}

				if currentReportWhen == "" || currentReportWhen != expectedReportWhen {
					targetSettings["reportWhen"] = expectedReportWhen
				}
			}
		}
		if v, ok := settingsSchemaMap["report_recipients"]; ok {
			if recipients, okList := v.([]any); okList && hasChanged("report_recipients", "reportRecipients", recipients) {
				targetSettings["reportRecipients"] = expandStringList(recipients)
			}
		}
		if v, ok := settingsSchemaMap["timezone"]; ok {
			if tzStr, okStr := v.(string); okStr && tzStr != "" && hasChanged("timezone", "timezone", tzStr) {
				targetSettings["timezone"] = tzStr
			}
		}
		if v, ok := settingsSchemaMap["export_retention"]; ok {
			if valInt, okInt := v.(int); okInt && hasChanged("export_retention", "exportRetention", valInt) {
				targetSettings["exportRetention"] = valInt
			}
		}
		if v, ok := settingsSchemaMap["delete_first"]; ok {
			if valBool, okBool := v.(bool); okBool && hasChanged("delete_first", "deleteFirst", valBool) {
				targetSettings["deleteFirst"] = valBool
			}
		}

		settingsMap[settingsKey] = targetSettings
	}

	if settingsMap[""] == nil {
		settingsMap[""] = defaultSettings
	}

	for key, value := range currentSettingsMap {
		if settingsMap[key] == nil {
			settingsMap[key] = value
		}
	}

	return settingsMap
}

func equalValues(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	switch aVal := a.(type) {
	case int:
		switch bVal := b.(type) {
		case int:
			return aVal == bVal
		case float64:
			return float64(aVal) == bVal
		}
	case float64:
		switch bVal := b.(type) {
		case int:
			return aVal == float64(bVal)
		case float64:
			return aVal == bVal
		}
	case bool:
		if bVal, ok := b.(bool); ok {
			return aVal == bVal
		}
	case string:
		if bVal, ok := b.(string); ok {
			return aVal == bVal
		}
	case []any:
		if bVal, ok := b.([]interface{}); ok {
			if len(aVal) != len(bVal) {
				return false
			}
			for i, item := range aVal {
				if !equalValues(item, bVal[i]) {
					return false
				}
			}
			return true
		}
	}

	return false
}

func parseBackupJobSettings(settingsMap map[string]any) map[string]any {
	result := make(map[string]any)

	if defaultSettings, ok := settingsMap[""].(map[string]any); ok {
		if val, exists := defaultSettings["reportWhen"]; exists {
			result["report_when_fail_only"] = (val.(string) == "failure")
		}
		if val, exists := defaultSettings["timezone"]; exists {
			result["timezone"] = val.(string)
		}

		if val, exists := defaultSettings["offlineBackup"]; exists {
			result["offline_backup"] = val.(bool)
		}
		if val, exists := defaultSettings["checkpointSnapshot"]; exists {
			result["checkpoint_snapshot"] = val.(bool)
		}
		if val, exists := defaultSettings["deleteFirst"]; exists {
			result["delete_first"] = val.(bool)
		}
		if val, exists := defaultSettings["remoteEnabled"]; exists {
			result["remote_enabled"] = val.(bool)
		}
		if val, exists := defaultSettings["compressionEnabled"]; exists {
			result["compression_enabled"] = val.(bool)
		}

		if val, exists := defaultSettings["retention"]; exists {
			if floatVal, ok := val.(float64); ok {
				result["retention"] = int(floatVal)
			} else if intVal, ok := val.(int); ok {
				result["retention"] = intVal
			}
		}
		if val, exists := defaultSettings["remoteRetention"]; exists {
			if floatVal, ok := val.(float64); ok {
				result["remote_retention"] = int(floatVal)
			} else if intVal, ok := val.(int); ok {
				result["remote_retention"] = intVal
			}
		}

		if val, exists := defaultSettings["reportRecipients"]; exists {
			if recipients, ok := val.([]interface{}); ok {
				strRecipients := make([]string, len(recipients))
				for i, r := range recipients {
					strRecipients[i] = r.(string)
				}
				result["report_recipients"] = strRecipients
			}
		}
	}

	for key, value := range settingsMap {
		if key != "" {
			if settingsMap, ok := value.(map[string]any); ok {
				if exportRetention, exists := settingsMap["exportRetention"]; exists {
					if floatVal, ok := exportRetention.(float64); ok {
						result["export_retention"] = int(floatVal)
					} else if intVal, ok := exportRetention.(int); ok {
						result["export_retention"] = intVal
					}
					if key != "" {
						result["schedule_id"] = key
					}
				}
			}
		}
	}

	return result
}

func convertMapToBackupSettings(settingsMap map[string]any) payloads.BackupSettings {
	settings := payloads.BackupSettings{}

	if defaultSettings, ok := settingsMap[""].(map[string]any); ok {
		if val, exists := defaultSettings["retention"]; exists {
			if intVal, ok := val.(int); ok {
				settings.Retention = &intVal
			}
		}

		if val, exists := defaultSettings["compressionEnabled"]; exists {
			if boolVal, ok := val.(bool); ok {
				settings.CompressionEnabled = &boolVal
			}
		}

		if val, exists := defaultSettings["offlineBackup"]; exists {
			if boolVal, ok := val.(bool); ok {
				settings.OfflineBackup = &boolVal
			}
		}
		if val, exists := defaultSettings["checkpointSnapshot"]; exists {
			if boolVal, ok := val.(bool); ok {
				settings.CheckpointSnapshot = &boolVal
			}
		}
		if val, exists := defaultSettings["remoteEnabled"]; exists {
			if boolVal, ok := val.(bool); ok {
				settings.RemoteEnabled = &boolVal
			}
		}
		if val, exists := defaultSettings["remoteRetention"]; exists {
			if intVal, ok := val.(int); ok {
				settings.RemoteRetention = &intVal
			}
		}
		if val, exists := defaultSettings["reportWhen"]; exists {
			if strVal, ok := val.(string); ok {
				reportWhen := payloads.ReportWhenAlways
				if strVal == "failure" {
					reportWhen = payloads.ReportWhenFailOnly
				}
				settings.ReportWhen = &reportWhen
			}
		}
		if val, exists := defaultSettings["reportRecipients"]; exists {
			if recipients, ok := val.([]string); ok {
				settings.ReportRecipients = recipients
			}
		}
		if val, exists := defaultSettings["timezone"]; exists {
			if strVal, ok := val.(string); ok {
				settings.Timezone = &strVal
			}
		}
		if val, exists := defaultSettings["deleteFirst"]; exists {
			if boolVal, ok := val.(bool); ok {
				settings.DeleteFirst = &boolVal
			}
		}
	}

	return settings
}
