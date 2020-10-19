package client

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type ResourceSet struct {
	Id       string            `json:"id"`
	Limits   ResourceSetLimits `json:"limits"`
	Name     string            `json:"name"`
	Subjects []string          `json:"subjects"`
	Objects  []string          `json:"objects"`
}

type ResourceSetLimits struct {
	Cpus   ResourceSetLimit `json:"cpus,omitempty"`
	Memory ResourceSetLimit `json:"memory,omitempty"`
	Disk   ResourceSetLimit `json:"disk,omitempty"`
}

type ResourceSetLimit struct {
	Available int `json:"available,omitempty"`
	Total     int `json:"total,omitempty"`
}

func (rs ResourceSet) Compare(obj interface{}) bool {
	other := obj.(ResourceSet)
	if other.Id == rs.Id {
		return true
	}

	if other.Name == rs.Name {
		return true
	}

	return false
}

func (c Client) GetResourceSets() ([]ResourceSet, error) {
	return c.makeResourceSetGetAllCall()
}

func (c Client) GetResourceSetById(id string) (*ResourceSet, error) {
	resourceSets, err := c.GetResourceSet(ResourceSet{
		Id: id,
	})

	if err != nil {
		return nil, err
	}

	l := len(resourceSets)
	if l != 1 {
		return nil, errors.New(fmt.Sprintf("found %d resource set(s) with id `%s`: %v", l, id, resourceSets))
	}

	return &resourceSets[0], nil
}

func (c Client) GetResourceSet(rsReq ResourceSet) ([]ResourceSet, error) {
	resourceSets, err := c.makeResourceSetGetAllCall()

	if err != nil {
		return nil, err
	}
	rsRv := []ResourceSet{}
	found := false
	for _, rs := range resourceSets {
		if rs.Compare(rsReq) {
			rsRv = append(rsRv, rs)
			found = true
		}
	}

	if !found {
		return rsRv, NotFound{Query: rsReq}
	}

	return rsRv, nil
}

func (c Client) makeResourceSetGetAllCall() ([]ResourceSet, error) {

	var res struct {
		ResourceSets []ResourceSet `json:"-"`
	}
	params := map[string]interface{}{
		"id": "dummy",
	}
	err := c.Call("resourceSet.getAll", params, &res.ResourceSets)
	log.Printf("[DEBUG] Calling resourceSet.getAll received response: %+v with error: %v\n", res, err)

	if err != nil {
		return nil, err
	}

	return res.ResourceSets, nil
}

func createLimitsMap(rsl ResourceSetLimits) map[string]interface{} {
	rv := map[string]interface{}{}

	if rsl.Cpus.Total != 0 {
		rv["cpus"] = rsl.Cpus
	}
	if rsl.Disk.Total != 0 {
		rv["disk"] = rsl.Disk
	}
	if rsl.Memory.Total != 0 {
		rv["memory"] = rsl.Memory
	}
	return rv
}

func (c Client) CreateResourceSet(rsReq ResourceSet) (*ResourceSet, error) {
	rs := ResourceSet{}
	limits := createLimitsMap(rsReq.Limits)
	params := map[string]interface{}{
		"name":     rsReq.Name,
		"subjects": rsReq.Subjects,
		"objects":  rsReq.Objects,
		"limits":   limits,
	}
	err := c.Call("resourceSet.create", params, &rs)
	log.Printf("[DEBUG] Calling resourceSet.create with params: %v returned: %+v with error: %v\n", params, rs, err)

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

		if len(rs) > 1 {
			return errors.New(fmt.Sprintf("refusing to delete resource set since `%d` resource sets were returned: %v", len(rs), rs))
		}

		id = rs[0].Id
	}
	var success bool
	params := map[string]interface{}{
		"id": id,
	}
	err := c.Call("resourceSet.delete", params, &success)
	log.Printf("[DEBUG] Calling resourceSet.delete call successful: %t with error: %v\n", success, err)

	return err
}

func (c Client) RemoveResourceSetSubject(rsReq ResourceSet, subject string) error {
	params := map[string]interface{}{
		"id":      rsReq.Id,
		"subject": subject,
	}
	var success bool
	err := c.Call("resourceSet.removeSubject", params, &success)
	log.Printf("[DEBUG] Calling resourceSet.removeSubject call successful: %t with error: %v\n", success, err)
	return err
}

func (c Client) AddResourceSetSubject(rsReq ResourceSet, subject string) error {
	params := map[string]interface{}{
		"id":      rsReq.Id,
		"subject": subject,
	}
	var success bool
	err := c.Call("resourceSet.addSubject", params, &success)
	log.Printf("[DEBUG] Calling resourceSet.addSubject call successful: %t with error: %v\n", success, err)
	return err
}

func (c Client) RemoveResourceSetObject(rsReq ResourceSet, object string) error {
	params := map[string]interface{}{
		"id":     rsReq.Id,
		"object": object,
	}
	var success bool
	err := c.Call("resourceSet.removeObject", params, &success)
	log.Printf("[DEBUG] Calling resourceSet.removeObject call successful: %t with error: %v\n", success, err)
	return err
}

func (c Client) AddResourceSetObject(rsReq ResourceSet, object string) error {
	params := map[string]interface{}{
		"id":     rsReq.Id,
		"object": object,
	}
	var success bool
	err := c.Call("resourceSet.addObject", params, &success)
	log.Printf("[DEBUG] Calling resourceSet.addObject call successful: %t with error: %v\n", success, err)
	return err
}

func (c Client) RemoveResourceSetLimit(rsReq ResourceSet, limit string) error {
	params := map[string]interface{}{
		"id":      rsReq.Id,
		"limitId": limit,
	}
	var success bool
	err := c.Call("resourceSet.removeLimit", params, &success)
	log.Printf("[DEBUG] Calling resourceSet.removeLimit call successful: %t with error: %v\n", success, err)
	return err
}

func (c Client) AddResourceSetLimit(rsReq ResourceSet, limit string, quantity int) error {
	params := map[string]interface{}{
		"id":       rsReq.Id,
		"limitId":  limit,
		"quantity": quantity,
	}
	var success bool
	err := c.Call("resourceSet.addLimit", params, &success)
	log.Printf("[DEBUG] Calling resourceSet.addLimit call with params: %v successful: %t with error: %v\n", params, success, err)
	return err
}

func RemoveResourceSetsWithNamePrefix(rsNamePrefix string) func(string) error {
	return func(_ string) error {
		fmt.Println("[DEBUG] Running sweeper")
		c, err := NewClient(GetConfigFromEnv())
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		rss, err := c.GetResourceSets()
		if err != nil {
			return fmt.Errorf("error getting resource sets: %s", err)
		}
		for _, rs := range rss {
			if strings.HasPrefix(rs.Name, rsNamePrefix) {
				err := c.DeleteResourceSet(ResourceSet{Id: rs.Id})

				if err != nil {
					log.Printf("error destroying resource set `%s` during sweep: %s", rs.Name, err)
				}
			}
		}
		return nil
	}
}
