package xoa

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func resourceTags() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		Elem: &schema.Schema{
			Type: schema.TypeString,
		},
	}
}

type ListTags []string

func New(i interface{}) ListTags {
	switch tags := i.(type) {
	case []interface{}:
		t := make(ListTags, len(tags))
		for _, tag := range tags {
			t = append(t, tag.(string))
		}
		return t
	default:
		return ListTags{}
	}
}

func (tags ListTags) Without(other ListTags) ListTags {
	remaining := ListTags{}

	for _, oldTag := range tags {
		found := false
		for _, newTag := range other {
			if oldTag == newTag {
				found = true
			}
		}

		if !found {
			remaining = append(remaining, oldTag)
		}
	}
	return remaining
}
