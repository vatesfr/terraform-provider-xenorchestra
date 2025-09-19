package xoa

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/vatesfr/xenorchestra-go-sdk/client"
)

func dataSourceXoaUser() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceUserReadContext,
		Description: "Provides information about a Xen Orchestra user. If the Xen Orchestra user account you are using is not an admin, see the `search_in_session` parameter.",
		Schema: map[string]*schema.Schema{
			"username": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The username of the XO user.",
				Required:    true,
			},
			"search_in_session": &schema.Schema{
				Type:        schema.TypeBool,
				Description: "A boolean which will search for the user in the current session (`session.getUser` Xen Orchestra RPC call). This allows a non admin user to look up their own user account.",
				Optional:    true,
				Default:     false,
			},
		},
	}
}

func dataSourceUserReadContext(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
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
		return diag.FromErr(err)
	}

	tflog.Debug(ctx, "Found user", map[string]interface{}{
		"user": *user,
	})

	d.SetId(user.Id)
	if err := d.Set("username", user.Email); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
