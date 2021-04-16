package internal

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func Test_TestCheckTypeListSorted(t *testing.T) {
	ascAttrs := map[string]string{
		"hosts.0.name_label": "A",
		"hosts.1.name_label": "B",
		"hosts.2.name_label": "C",
	}
	descAttrs := map[string]string{
		"hosts.0.name_label": "C",
		"hosts.1.name_label": "B",
		"hosts.2.name_label": "A",
	}
	tests := []struct {
		name           string
		tfstate        *terraform.State
		expectedErrMsg string
		sortOrder      string
	}{
		{
			name:           "identifyIncorrectDescSorting",
			expectedErrMsg: "to be sorted",
			sortOrder:      sortOrderDesc,
			tfstate:        getTfstate(ascAttrs),
		},
		{
			name:           "identifyCorrectDescSorting",
			expectedErrMsg: "",
			sortOrder:      sortOrderDesc,
			tfstate:        getTfstate(descAttrs),
		},
		{
			name:           "identifyIncorrectAscSorting",
			expectedErrMsg: "to be sorted",
			sortOrder:      sortOrderAsc,
			tfstate:        getTfstate(descAttrs),
		},
		{
			name:           "identifyCorrectAscSorting",
			expectedErrMsg: "",
			sortOrder:      sortOrderAsc,
			tfstate:        getTfstate(ascAttrs),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := TestCheckTypeListAttrSorted("resource", "hosts.*.name_label", tt.sortOrder)(tt.tfstate)
			if tt.expectedErrMsg == "" && err != nil {
				t.Fatalf("expected result to be nil but received %v", err)
			}

			if tt.expectedErrMsg != "" && err == nil {
				t.Fatalf("expected %s but instead received nil error", tt.expectedErrMsg)
			}

			if tt.expectedErrMsg != "" && !strings.Contains(err.Error(), tt.expectedErrMsg) {
				t.Fatalf("expected error %v to contain '%s'", err, tt.expectedErrMsg)
			}
		})
	}
}

func getTfstate(attrs map[string]string) *terraform.State {
	return &terraform.State{
		Modules: []*terraform.ModuleState{
			&terraform.ModuleState{
				Path: []string{"root"},
				Resources: map[string]*terraform.ResourceState{
					"resource": &terraform.ResourceState{
						Primary: &terraform.InstanceState{
							Attributes: attrs,
						},
					},
				},
			},
		},
	}

}
