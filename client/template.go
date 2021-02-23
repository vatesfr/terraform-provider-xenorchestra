package client

import (
	"errors"
	"fmt"
	"os"
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

func FindTemplateForTests(template *Template, poolId string) {
	var found bool
	templateName, found := os.LookupEnv("XOA_TEMPLATE")
	if !found {
		fmt.Println("The XOA_TEMPLATE environment variable must be set for the tests")
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
