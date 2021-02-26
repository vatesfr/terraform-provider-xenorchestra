package client

import (
	"testing"
)

func TestGetTemplate(t *testing.T) {
	tests := []struct {
		templateName string
		template     Template
		err          error
	}{
		{
			templateName: testTemplate.NameLabel,
			template: Template{
				NameLabel: testTemplate.NameLabel,
			},
			err: nil,
		},
		{
			templateName: "Not found",
			template:     Template{},
			err:          NotFound{Query: Template{NameLabel: "Not found"}},
		},
	}

	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Errorf("failed to create client: %v", err)
	}

	for _, test := range tests {

		templateName := test.templateName
		templates, err := c.GetTemplate(Template{NameLabel: templateName})

		failureMsg := "failed to get template `%s` expected err: %v received: %v"
		if test.err == nil {
			if test.err != err {
				t.Fatalf(failureMsg, templateName, test.err, err)
			}

		} else {
			if test.err.Error() != err.Error() {
				t.Fatalf(failureMsg, templateName, test.err, err)
			}
		}

		if _, ok := test.err.(NotFound); ok {
			continue
		}

		if len(templates) < 1 {
			t.Fatalf("failed to find templates for the following name_label: %s", templateName)
		}
		tmp := templates[0]
		if test.template.NameLabel != "" && tmp.NameLabel != templateName {
			t.Errorf("template returned from xoa does not match. expected %s, found %s", tmp.NameLabel, templateName)
		}
	}
}
