package xoa

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/vatesfr/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var (
	retryModeMap = map[string]client.RetryMode{
		"none":    client.None,
		"backoff": client.Backoff,
	}
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("XOA_URL", nil),
				Description: "Hostname of the xoa router. Can be set via the XOA_URL environment variable.",
			},
			"username": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("XOA_USER", nil),
				Description:   "User account for xoa api. Can be set via the XOA_USER environment variable.",
				RequiredWith:  []string{"password"},
				ConflictsWith: []string{"token"},
			},
			"password": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("XOA_PASSWORD", nil),
				Description:   "Password for xoa api. Can be set via the XOA_PASSWORD environment variable.",
				RequiredWith:  []string{"username"},
				ConflictsWith: []string{"token"},
			},
			"token": &schema.Schema{
				Type:          schema.TypeString,
				Optional:      true,
				DefaultFunc:   schema.EnvDefaultFunc("XOA_TOKEN", nil),
				Description:   "Password for xoa api. Can be set via the XOA_TOKEN environment variable.",
				ConflictsWith: []string{"username", "password"},
			},
			"insecure": &schema.Schema{
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				DefaultFunc: schema.EnvDefaultFunc("XOA_INSECURE", nil),
				Description: "Whether SSL should be verified or not. Can be set via the XOA_INSECURE environment variable.",
			},
			"retry_mode": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("XOA_RETRY_MODE", "backoff"),
				Description:  "Specifies if retries should be attempted for requests that require eventual . Can be set via the XOA_RETRY_MODE environment variable.",
				ValidateFunc: validation.StringInSlice([]string{"backoff", "none"}, false),
			},
			"retry_max_time": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				DefaultFunc:  schema.EnvDefaultFunc("XOA_RETRY_MAX_TIME", "5m"),
				Description:  "If `retry_mode` is set, this specifies the duration for which the backoff method will continue retries. Can be set via the `XOA_RETRY_MAX_TIME` environment variable",
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^[0-9]+(\.[0-9]+)?(ms|s|m|h)$`), "must be a number immediately followed by ms (milliseconds), s (seconds), m (minutes), or h (hours). For example, \"30s\" for 30 seconds."),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"xenorchestra_acl":            resourceAcl(),
			"xenorchestra_bonded_network": resourceXoaBondedNetwork(),
			"xenorchestra_cloud_config":   resourceCloudConfigRecord(),
			"xenorchestra_network":        resourceXoaNetwork(),
			"xenorchestra_vm":             resourceRecord(),
			"xenorchestra_resource_set":   resourceResourceSet(),
			"xenorchestra_vdi":            resourceVDIRecord(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"xenorchestra_cloud_config": dataSourceXoaCloudConfig(),
			"xenorchestra_network":      dataSourceXoaNetwork(),
			"xenorchestra_pif":          dataSourceXoaPIF(),
			"xenorchestra_pool":         dataSourceXoaPool(),
			"xenorchestra_host":         dataSourceXoaHost(),
			"xenorchestra_hosts":        dataSourceXoaHosts(),
			"xenorchestra_template":     dataSourceXoaTemplate(),
			"xenorchestra_resource_set": dataSourceXoaResourceSet(),
			"xenorchestra_sr":           dataSourceXoaStorageRepository(),
			"xenorchestra_user":         dataSourceXoaUser(),
			"xenorchestra_vms":          dataSourceXoaVms(),
			"xenorchestra_vdi":          dataSourceXoaVDI(),
		},
		ConfigureFunc: xoaConfigure,
	}
}

func xoaConfigure(d *schema.ResourceData) (interface{}, error) {
	url := d.Get("url").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	token := d.Get("token").(string)
	insecure := d.Get("insecure").(bool)
	retryMode := d.Get("retry_mode").(string)
	retryMaxTime := d.Get("retry_max_time").(string)

	duration, err := time.ParseDuration(retryMaxTime)
	if err != nil {
		return client.Config{}, err
	}

	retry, ok := retryModeMap[retryMode]
	if !ok {
		return client.Config{}, errors.New(fmt.Sprintf("retry mode provided invalid: %s", retryMode))
	}

	config := client.Config{
		Url:                url,
		Username:           username,
		Password:           password,
		Token:              token,
		InsecureSkipVerify: insecure,
		RetryMode:          retry,
		RetryMaxTime:       duration,
	}
	return client.NewClient(config)
}
