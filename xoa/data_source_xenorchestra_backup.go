package xoa

import (
	"context"
	"fmt"

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
				Description: "The ID of the backup job. Required if `name` is not set.",
				Computed:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "The name of the backup job. Required if `id` is not set.",
				Computed:    true,
			},
			"mode": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The mode of the backup job.",
			},
			"schedule": {
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The schedule configuration for the backup job.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cron": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The cron expression for the backup job schedule.",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the backup job schedule is enabled.",
						},
						"timezone": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The timezone for the backup job schedule.",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "The name for the backup job schedule.",
						},
					},
				},
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
				Type:        schema.TypeSet,
				Computed:    true,
				Description: "The settings for the backup job.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"retention": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"compression_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"offline_backup": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"checkpoint_snapshot": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"remote_enabled": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"remote_retention": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"report_when_fail_only": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"report_recipients": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"timezone": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"export_retention": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"delete_first": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"backup_report_tpl": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"merge_backups_synchronously": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"max_export_rate": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"concurrency": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"long_term_retention": {
							Type:     schema.TypeList,
							Computed: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"daily": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"retention": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"weekly": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"retention": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"monthly": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"retention": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
									"yearly": {
										Type:     schema.TypeList,
										Computed: true,
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"retention": {
													Type:     schema.TypeInt,
													Computed: true,
												},
											},
										},
									},
								},
							},
						},
						"offline_snapshot": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"copy_retention": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"cbt_destroy_snapshot_data": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"nbd_concurrency": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"prefer_nbd": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"retention_pool_metadata": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"retention_xo_metadata": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"timeout": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"n_retries_vm_backup_failures": {
							Type:     schema.TypeInt,
							Computed: true,
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
	}
}

func dataSourceBackupRead(d *schema.ResourceData, m any) error {
	c := m.(*v2.XOClient)
	ctx := context.Background()

	jobID := d.Get("id").(string)
	jobName := d.Get("name").(string)

	var jobPayload *payloads.BackupJobResponse
	var err error

	if jobID != "" {
		jobPayload, err = c.Backup().GetJob(ctx, jobID, payloads.RestAPIJobQueryVM)
		if err != nil {
			d.SetId("")
			return nil
		}
	} else if jobName != "" {
		jobs, listErr := c.Backup().ListJobs(ctx, 0, payloads.RestAPIJobQueryVM)
		if listErr != nil {
			return fmt.Errorf("failed to list backup jobs: %w", listErr)
		}
		found := false
		for _, job := range jobs {
			if job.Name == jobName {
				jobPayload = job
				found = true
				break
			}
		}
		if !found {
			d.SetId("")
			return nil
		}
	} else {
		return fmt.Errorf("either id or name must be specified")
	}

	d.SetId(jobPayload.ID.String())

	// Set basic job properties
	d.Set("name", jobPayload.Name)
	d.Set("mode", string(jobPayload.Mode))

	// Process VMs
	var vmList []string
	switch vms := jobPayload.VMs.(type) {
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
	case map[string]any:
		vmsMap := vms
		if idVal, ok := vmsMap["id"]; ok {
			if idStr, ok := idVal.(string); ok && idStr != "" {
				vmList = []string{idStr}
			}
		}
	}
	d.Set("vms", vmList)

	// Parse settings from new BackupJobResponse format
	parsedSettings := parseBackupJobSettingsForDataSource(jobPayload.Settings)

	if len(parsedSettings) > 0 {
		d.Set("settings", []any{parsedSettings})
	}

	schedules, err := c.Schedule().GetAll(ctx)
	if err != nil {
		return fmt.Errorf("failed to get schedules: %w", err)
	}

	for _, schedule := range schedules {
		if schedule.JobID == jobPayload.ID {
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

	d.Set("schedule", []any{})

	return nil
}

func parseBackupJobSettingsForDataSource(settingsMap map[string]any) map[string]any {
	result := make(map[string]any)

	if defaultSettings, ok := settingsMap[""].(map[string]any); ok {
		// Helper function: only set field if it exists in API
		parseOptionalField := func(apiKey, tfKey string, isInt bool) {
			if val, exists := defaultSettings[apiKey]; exists {
				if isInt {
					if floatVal, ok := val.(float64); ok {
						result[tfKey] = int(floatVal)
					} else if intVal, ok := val.(int); ok {
						result[tfKey] = intVal
					}
				} else {
					result[tfKey] = val
				}
			}
		}

		// Field mappings
		intFields := map[string]string{
			"retention":                "retention",
			"remoteRetention":          "remote_retention",
			"copyRetention":            "copy_retention",
			"concurrency":              "concurrency",
			"maxExportRate":            "max_export_rate",
			"nRetriesVmBackupFailures": "n_retries_vm_backup_failures",
			"nbdConcurrency":           "nbd_concurrency",
			"timeout":                  "timeout",
			"retentionPoolMetadata":    "retention_pool_metadata",
			"retentionXoMetadata":      "retention_xo_metadata",
		}

		boolFields := map[string]string{
			"offlineBackup":             "offline_backup",
			"offlineSnapshot":           "offline_snapshot",
			"checkpointSnapshot":        "checkpoint_snapshot",
			"deleteFirst":               "delete_first",
			"remoteEnabled":             "remote_enabled",
			"compressionEnabled":        "compression_enabled",
			"mergeBackupsSynchronously": "merge_backups_synchronously",
			"cbtDestroySnapshotData":    "cbt_destroy_snapshot_data",
			"preferNbd":                 "prefer_nbd",
		}

		stringFields := map[string]string{
			"timezone":        "timezone",
			"backupReportTpl": "backup_report_tpl",
		}

		// Process all field types
		for apiKey, tfKey := range intFields {
			parseOptionalField(apiKey, tfKey, true)
		}
		for apiKey, tfKey := range boolFields {
			parseOptionalField(apiKey, tfKey, false)
		}
		for apiKey, tfKey := range stringFields {
			parseOptionalField(apiKey, tfKey, false)
		}

		// Special cases
		if val, exists := defaultSettings["reportWhen"]; exists {
			if strVal, ok := val.(string); ok {
				result["report_when_fail_only"] = (strVal == "failure")
			}
		}

		if val, exists := defaultSettings["reportRecipients"]; exists {
			if recipients, ok := val.([]interface{}); ok {
				strRecipients := make([]string, len(recipients))
				for i, r := range recipients {
					if str, ok := r.(string); ok {
						strRecipients[i] = str
					}
				}
				result["report_recipients"] = strRecipients
			}
		}

		// Long-term retention
		if val, exists := defaultSettings["longTermRetention"]; exists {
			if ltrMap, ok := val.(map[string]any); ok {
				longTermRetention := make([]map[string]any, 1)
				ltrData := make(map[string]any)

				for _, period := range []string{"daily", "weekly", "monthly", "yearly"} {
					if periodData, exists := ltrMap[period]; exists {
						if periodMap, ok := periodData.(map[string]any); ok {
							if retention, hasRetention := periodMap["retention"]; hasRetention {
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

	// Handle schedule-specific settings
	for key, value := range settingsMap {
		if key != "" {
			if settingsData, ok := value.(map[string]any); ok {
				if exportRetention, exists := settingsData["exportRetention"]; exists {
					if floatVal, ok := exportRetention.(float64); ok {
						result["export_retention"] = int(floatVal)
					} else if intVal, ok := exportRetention.(int); ok {
						result["export_retention"] = intVal
					}
				}
			}
		}
	}

	return result
}
