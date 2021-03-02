package client

import (
	"errors"
	"fmt"
	"os"
)

type TemplateDisk struct {
	Bootable bool   `json:"bootable"`
	Device   string `json:"device"`
	Size     int    `json:"size"`
	Type     string `json:"type"`
	SR       string `json:"SR"`
}

type TemplateInfo struct {
	Arch  string         `json:"arch"`
	Disks []TemplateDisk `json:"disks"`
}

type Template struct {
	Id           string       `json:"id"`
	Uuid         string       `json:"uuid"`
	NameLabel    string       `json:"name_label"`
	PoolId       string       `json:"$poolId"`
	TemplateInfo TemplateInfo `json:"template_info"`
}

func (t Template) Compare(obj interface{}) bool {
	other := obj.(Template)

	if t.Id == other.Id {
		return true
	}

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

func (t Template) isDiskTemplate() bool {
	if len(t.TemplateInfo.Disks) == 0 && t.NameLabel != "Other install media" {
		return true
	}

	return false
}

func (c *Client) GetTemplate(template Template) ([]Template, error) {
	obj, err := c.FindFromGetAllObjects(template)
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

func FindTemplateForTests(template *Template, poolId, templateEnvVar string) {
	var found bool
	templateName, found := os.LookupEnv(templateEnvVar)
	if !found {
		fmt.Println(fmt.Sprintf("The %s environment variable must be set for the tests", templateEnvVar))
		os.Exit(-1)
	}

	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		fmt.Printf("failed to create client with error: %v", err)
		os.Exit(-1)
	}

	templates, err := c.GetTemplate(Template{
		NameLabel: templateName,
		PoolId:    poolId,
	})

	if err != nil {
		fmt.Printf("failed to find templates with error: %v\n", err)
		os.Exit(-1)
	}

	l := len(templates)
	if l != 1 {
		fmt.Printf("found %d templates when expected to find 1. templates found: %v\n", l, templates)
		os.Exit(-1)
	}
	*template = templates[0]
}
