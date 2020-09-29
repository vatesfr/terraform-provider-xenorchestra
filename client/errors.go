package client

import (
	"fmt"
)

type NotFound struct {
	Query XoObject
}

func (e NotFound) Error() string {
	return fmt.Sprintf("Could not find %[1]T with query: %+[1]v", e.Query)
}
