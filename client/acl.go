package client

import (
	"errors"
	"log"
)

type Acl struct {
	Id      string
	Action  string
	Subject string
	Object  string
}

func (acl Acl) Compare(obj interface{}) bool {
	other := obj.(Acl)

	if acl.Id == other.Id {
		return true
	}

	if acl.Action == other.Action && acl.Subject == other.Subject && acl.Object == other.Object {
		return true
	}

	return false
}

func (c *Client) CreateAcl(acl Acl) (*Acl, error) {
	var success bool
	params := map[string]interface{}{
		"subject": acl.Subject,
		"object":  acl.Object,
		"action":  acl.Action,
	}
	err := c.Call("acl.add", params, &success)

	if err != nil {
		return nil, err
	}

	return c.GetAcl(acl)
}

func (c *Client) GetAcls() ([]Acl, error) {
	params := map[string]interface{}{
		"dummy": "dummy",
	}
	acls := []Acl{}
	err := c.Call("acl.get", params, &acls)

	if err != nil {
		return nil, err
	}
	log.Printf("[DEBUG] Found the following ACLs: %v\n", acls)
	return acls, nil
}

func (c *Client) GetAcl(aclReq Acl) (*Acl, error) {
	acls, err := c.GetAcls()
	if err != nil {
		return nil, err
	}

	var foundAcl Acl
	for _, acl := range acls {
		if acl.Compare(aclReq) {
			foundAcl = acl
		}
	}

	if foundAcl.Id == "" {
		return nil, NotFound{Query: aclReq}
	}

	return &foundAcl, nil
}

func (c *Client) DeleteAcl(acl Acl) error {
	var err error
	var aclRef *Acl
	if getAclById(acl) {
		aclRef, err = c.GetAcl(acl)
		acl = *aclRef
	}
	var success bool
	params := map[string]interface{}{
		"subject": acl.Subject,
		"object":  acl.Object,
		"action":  acl.Action,
	}
	err = c.Call("acl.remove", params, &success)

	if err != nil {
		return err
	}

	if !success {
		return errors.New("failed to delete acl")
	}
	return nil
}

func getAclById(acl Acl) bool {
	if acl.Id != "" && acl.Subject == "" && acl.Object == "" && acl.Action == "" {
		return true
	}

	return false
}
