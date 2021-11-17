package xoa

import (
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xo-sdk-go/client"
)

func dataSourceXoaUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func dataSourceUserRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	username := d.Get("username").(string)

	user, err := c.GetUser(client.User{
		Email: username,
	})

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Found user with %+v", user)

	d.SetId(user.Id)
	if err := d.Set("username", user.Email); err != nil {
		return err
	}

	return nil
}
