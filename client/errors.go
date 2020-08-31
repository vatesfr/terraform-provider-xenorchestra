package client

import "fmt"

type NotFound struct {
	Query XoObject
	Type  string
}

func (e NotFound) Error() string {
	return fmt.Sprintf("Could not find %s with query: %+v", e.Type, e.Query)
}
