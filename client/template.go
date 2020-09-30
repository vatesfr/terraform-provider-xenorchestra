package client

import "errors"

type Template struct {
	Id        string `json:"id"`
	Uuid      string `json:"uuid"`
	NameLabel string `json:"name_label"`
}

func (t Template) Compare(obj interface{}) bool {
	other := obj.(Template)
	if t.NameLabel == other.NameLabel {
		return true
	}
	return false
}

func (c *Client) GetTemplate(name string) ([]Template, error) {
	obj, err := c.FindFromGetAllObjects(Template{NameLabel: name})
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
