package client

import "fmt"

type NotFound struct {
	Id   string
	Type string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("Could not find %s with id: %s", e.Type, e.Id)
}
