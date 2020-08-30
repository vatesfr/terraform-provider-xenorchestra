package client

import "fmt"

type NotFound struct {
	Id   string
	Type string
}

func (e NotFound) Error() string {
	// TODO: This does not help when we query for a non ID value like name_label
	// See https://github.com/terra-farm/terraform-provider-xenorchestra/issues/58
	// for more details.
	return fmt.Sprintf("Could not find %s with id: %s", e.Type, e.Id)
}
