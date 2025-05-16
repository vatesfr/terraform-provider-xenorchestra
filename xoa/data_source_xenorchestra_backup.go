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

	var jobPayload *payloads.BackupJob
	var err error

	if jobID != "" {
		jobPayload, err = c.Backup().GetJob(ctx, jobID)
		if err != nil {
			d.SetId("")
			return nil
		}
	} else if jobName != "" {
		jobs, listErr := c.Backup().ListJobs(ctx, 0)
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
	case map[string]interface{}:
		vmsMap := vms
		if idVal, ok := vmsMap["id"]; ok {
			if idStr, ok := idVal.(string); ok && idStr != "" {
				vmList = []string{idStr}
			}
		}
	}
	d.Set("vms", vmList)

	// Process settings
	if jobPayload.Settings != nil {
		if settings, ok := jobPayload.Settings[""]; ok {
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
			if settings.Timezone != nil {
				settingsMap["timezone"] = *settings.Timezone
			}
			d.Set("settings", []any{settingsMap})
		}
	}

	// Get schedule information
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

	// If no schedule is found, set empty schedule
	d.Set("schedule", []any{})

	return nil
}
