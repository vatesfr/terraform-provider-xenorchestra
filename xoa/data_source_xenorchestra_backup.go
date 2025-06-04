package xoa

import (
	"context"
	"fmt"

	"github.com/gofrs/uuid"
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
						"export_retention": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"delete_first": {
							Type:     schema.TypeBool,
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
						"backup_report_tpl": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"merge_backups_synchronously": {
							Type:     schema.TypeBool,
							Computed: true,
						},
						"concurrency": {
							Type:     schema.TypeInt,
							Computed: true,
						},
						"max_export_rate": {
							Type:     schema.TypeInt,
							Computed: true,
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

	var job *payloads.BackupJobResponse
	var err error

	if jobID != "" {
		job, err = c.Backup().GetJob(ctx, jobID, payloads.RestAPIJobQueryVM)
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
		for _, j := range jobs {
			if j.Name == jobName {
				job = j
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

	d.SetId(job.ID.String())
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
