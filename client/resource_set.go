package client

import (
	"context"
	"fmt"
	"time"
)

type ResourceSet struct {
	Id       string            `json:"id"`
	Limits   ResourceSetLimits `json:"limits"`
	Name     string            `json:"name"`
	Subjects []string          `json:"subjects"`
	Objects  []string          `json:"objects"`
}

type ResourceSetLimits struct {
	Cpus   ResourceSetLimit
	Memory ResourceSetLimit
	Disk   ResourceSetLimit
}

type ResourceSetLimit struct {
	Available int
	Total     int
}

func (rs ResourceSet) New(obj map[string]interface{}) XoObject {

	id := obj["id"].(string)

	return ResourceSet{
		Id: id,
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
	}
}

func (rs ResourceSet) Compare(obj map[string]interface{}) bool {
	name := obj["name"].(string)

	if name == rs.Name {
		return true
	}

	return false
}

func (c Client) GetResourceSet(rsReq ResourceSet) (*ResourceSet, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	var res struct {
		ResourceSets []ResourceSet `json:"-"`
	}
	params := map[string]interface{}{
		"id": "dummy",
	}
	err := c.Call(ctx, "resourceSet.getAll", params, &res.ResourceSets)
	fmt.Printf("[DEBUG] Calling resourceSet.getAll received response: %+v with error: %v\n", res, err)

	if err != nil {
		return nil, err
	}

	found := false
	var rsRv *ResourceSet
	for _, rs := range res.ResourceSets {
		if rsReq.Name == rs.Name {
			found = true
			rsRv = &rs
		}
	}

	if !found {
		return rsRv, NotFound{}
	}

	return rsRv, nil
}

func (c Client) CreateResourceSet(rsReq ResourceSet) (*ResourceSet, error) {
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	rs := ResourceSet{}
	params := map[string]interface{}{
		"name":     rsReq.Name,
		"subjects": rsReq.Subjects,
		"objects":  rsReq.Objects,
		"limits":   rsReq.Limits,
	}
	err := c.Call(ctx, "resourceSet.create", params, &rs)
	fmt.Printf("[DEBUG] Calling resourceSet.create returned: %+v with error: %v\n", rs, err)

	if err != nil {
		return nil, err
	}

	return &rs, err
}

func (c Client) DeleteResourceSet(rsReq ResourceSet) error {

	id := rsReq.Id
	if id == "" {
		rs, err := c.GetResourceSet(rsReq)

		if err != nil {
			return err
		}

		id = rs.Id
	}
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	var success bool
	params := map[string]interface{}{
		"id": id,
	}
	err := c.Call(ctx, "resourceSet.delete", params, &success)
	fmt.Printf("[DEBUG] Calling resourceSet.delete call successful: %t with error: %v\n", success, err)

	return err
}
