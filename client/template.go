package client

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Template struct {
	// TODO: Not sure the difference between these two
	Id        string
	Uuid      string
	NameLabel string
}

func (c *Client) GetTemplate(name string) (Template, error) {
	var template Template
	params := map[string]interface{}{
		"type": "VM-template",
	}
	var objsRes struct {
		Templates map[string]interface{} `json:"-"`
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	err := c.rpc.Call(ctx, "xo.getAllObjects", params, &objsRes.Templates)

	if err != nil {
		return template, err
	}

	found := false
	for _, obj := range objsRes.Templates {
		v, ok := obj.(map[string]interface{})
		if !ok {
			return template, errors.New("Could not coerce interface{} into map")
		}

		if v["type"].(string) != "VM-template" {
			continue
		}

		name_label, ok := v["name_label"].(string)

		if !ok {
			return template, errors.New(fmt.Sprintf("type assertion for name_label failed on VM-template: %v", v))
		}

		if name_label == name {
			found = true
			// TODO: Add stricter error checking here
			id := v["id"].(string)
			uuid := v["uuid"].(string)
			template = Template{
				Id:        id,
				NameLabel: name_label,
				Uuid:      uuid,
			}
		}
	}

	if !found {
		return template, NotFound{Type: "VM-template"}
	}
	return template, nil
}
