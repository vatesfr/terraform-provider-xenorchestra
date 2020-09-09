package client

import (
	"testing"
)

var testResourceSetName string = "terraform-acc-resource-set2"

var testResourceSet = ResourceSet{
	Name: testResourceSetName,
	Limits: ResourceSetLimits{
		Cpus: ResourceSetLimit{
			Total:     1,
			Available: 2,
		},
		Disk: ResourceSetLimit{
			Total:     1,
			Available: 2,
		},
		Memory: ResourceSetLimit{
			Total:     1,
			Available: 2,
		},
	},
	Subjects: []string{},
	Objects:  []string{},
}

var resourceSetObj = map[string]interface{}{
	"id":   "id of resource set",
	"name": "resource set name",
	"limits": map[string]interface{}{
		"cpus": map[string]interface{}{
			"available": 4,
			"total":     4,
		},
		"disk": map[string]interface{}{
			"available": 4,
			"total":     4,
		},
		"memory": map[string]interface{}{
			"available": 4,
			"total":     4,
		},
	},
}

func TestResourceSetCompare(t *testing.T) {
	tests := []struct {
		object map[string]interface{}
		rs     ResourceSet
		result bool
	}{
		{
			object: resourceSetObj,
			rs: ResourceSet{
				Name: "resource set name",
			},
			result: true,
		},
		{
			object: resourceSetObj,
			rs: ResourceSet{
				Name: "Not the same name",
			},
			result: false,
		},
	}

	for _, test := range tests {
		rs := test.rs
		result := test.result
		object := test.object

		if rs.Compare(test.object) != result {
			t.Errorf("error: expected object `%v` to Compare `%t` to ResourceSet `%v`", object, result, rs)
		}
	}
}

func TestGetResourceSet(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	rs, err := c.GetResourceSet(ResourceSet{
		Name: testResourceSetName,
	})

	if err != nil {
		t.Fatalf("failed to retrieve ResourceSet with error: %v", err)
	}

	if rs.Name != testResourceSetName {
		t.Errorf("resource set's name `%s` did not match expected `%s`", rs.Name, testResourceSetName)
	}

	if rs.Limits.Cpus.Available != 2 {
		t.Errorf("resource set should have contained 2 CPUs")
	}
}
