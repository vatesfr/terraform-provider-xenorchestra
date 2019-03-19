package client

import (
	"testing"
)

func TestGetTemplate(t *testing.T) {
	// TODO: Export this function so it can be used by the client tests
	// xoa.testAccPreCheck(t)
	tests := []struct {
		templateName string
		template     Template
		err          error
	}{
		{
			templateName: "Asianux Server 4 (64-bit)",
			template: Template{
				NameLabel: "Asianux Server 4 (64-bit)",
			},
			err: nil,
		},
		{
			templateName: "No found",
			template:     Template{},
			err:          NotFound{Type: "VM-template"},
		},
	}

	c, err := NewClient()

	if err != nil {
		t.Errorf("failed to create client: %v", err)
	}

	for _, test := range tests {

		templateName := test.templateName
		tmp, err := c.GetTemplate(templateName)

		if test.err != err {
			t.Errorf("failed to get template `%s` expected err: %v received: %v", templateName, test.err, err)
		}

		if test.template.NameLabel != "" && tmp.NameLabel != templateName {
			t.Errorf("template returned from xoa does not match. expected %s, found %s", tmp.NameLabel, templateName)
		}
	}
}
