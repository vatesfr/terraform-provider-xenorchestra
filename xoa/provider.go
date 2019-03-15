package xoa

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("XOA_HOST", nil),
				Description: "Hostname of the xoa router",
			},
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("XOA_USER", nil),
				Description: "User account for xoa api",
			},
			"password": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("XOA_PASSWORD", nil),
				Description: "Password for xoa api",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"xenorchestra_vm": resourceRecord(),
		},
		// TODO: do i need a configure func?
		ConfigureFunc: xoaConfigure,
	}
}

func xoaConfigure(d *schema.ResourceData) (c interface{}, err error) {
	address := d.Get("host").(string)
	username := d.Get("username").(string)
	password := d.Get("password").(string)
	c = xoaConfig{
		host:     address,
		username: username,
		password: password,
	}
	return
}
