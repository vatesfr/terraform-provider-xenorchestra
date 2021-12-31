package xoa

import (
	"log"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataSourceXoaUser() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceUserRead,
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},
			"search_in_session": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func dataSourceUserRead(d *schema.ResourceData, m interface{}) error {
	c := m.(client.XOClient)

	username := d.Get("username").(string)
	searchInSession := d.Get("search_in_session").(bool)

	var user *client.User
	var err error
	if searchInSession {
		user, err = c.GetCurrentUser()
	} else {
		user, err = c.GetUser(client.User{
			Email: username,
		})
	}

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Found user with %+v", *user)

	d.SetId(user.Id)
	if err := d.Set("username", user.Email); err != nil {
		return err
	}

	return nil
}
