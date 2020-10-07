package client

import (
	"errors"
	"fmt"
)

type Template struct {
	Id        string `json:"id"`
	Uuid      string `json:"uuid"`
	NameLabel string `json:"name_label"`
	PoolId    string `json:"$poolId"`
}

func (t Template) Compare(obj interface{}) bool {
	other := obj.(Template)

	labelsMatch := false
	if t.NameLabel == other.NameLabel {
		labelsMatch = true
	}

	if t.PoolId == "" && labelsMatch {
		return true
	} else if t.PoolId == other.PoolId && labelsMatch {
		return true
	}
	return false
}

func (c *Client) GetTemplate(template Template) ([]Template, error) {
	obj, err := c.FindFromGetAllObjects(template)
	fmt.Println(fmt.Sprintf("template %v", template))
	var templates []Template
	if err != nil {
		return templates, err
	}

	templates, ok := obj.([]Template)

	if !ok {
		return templates, errors.New("failed to coerce response into Template slice")
	}

	return templates, nil
}
