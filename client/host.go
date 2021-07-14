package client

import (
	"errors"
	"fmt"
	"os"
	"sort"
)

type Host struct {
	Id        string           `json:"id"`
	NameLabel string           `json:"name_label"`
	Tags      []interface{}    `json:"tags,omitempty"`
	Pool      string           `json:"$pool"`
	Memory    HostMemoryObject `json:"memory"`
	Cpus      CpuInfo          `json:"cpus"`
}

type HostMemoryObject struct {
	Usage int `json:"usage"`
	Size  int `json:"size"`
}

func (h Host) Compare(obj interface{}) bool {
	otherHost := obj.(Host)
	if otherHost.Id == h.Id {
		return true
	}
	if h.Pool == otherHost.Pool {
		return true
	}
	if h.NameLabel != "" && h.NameLabel == otherHost.NameLabel {
		return true
	}
	return false
}

func (c *Client) GetHostByName(nameLabel string) (hosts []Host, err error) {
	obj, err := c.FindFromGetAllObjects(Host{NameLabel: nameLabel})
	if err != nil {
		return
	}
	return obj.([]Host), nil
}

func (c *Client) GetHostById(id string) (host Host, err error) {
	obj, err := c.FindFromGetAllObjects(Host{Id: id})
	if err != nil {
		return
	}
	hosts, ok := obj.([]Host)

	if !ok {
		return host, errors.New("failed to coerce response into Host slice")
	}

	if len(hosts) != 1 {
		return host, errors.New(fmt.Sprintf("expected a single host to be returned, instead received: %d in the response: %v", len(hosts), obj))
	}

	return hosts[0], nil
}

func FindHostForTests(hostId string, host *Host) {
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		fmt.Printf("failed to create client with error: %v", err)
		os.Exit(-1)
	}

	queriedHost, err := c.GetHostById(hostId)

	if err != nil {
		fmt.Printf("failed to find a host with id: %v with error: %v\n", hostId, err)
		os.Exit(-1)
	}

	*host = queriedHost
}

func (c *Client) GetSortedHosts(host Host, sortBy, sortOrder string) (hosts []Host, err error) {
	obj, err := c.FindFromGetAllObjects(host)

	if err != nil {
		return
	}
	slice := obj.([]Host)

	return sortHostsByField(slice, sortBy, sortOrder), nil
}

const (
	sortOrderAsc       = "asc"
	sortOrderDesc      = "desc"
	sortFieldId        = "id"
	sortFieldNameLabel = "name_label"
)

func sortHostsByField(hosts []Host, by, order string) []Host {
	if by == "" || order == "" {
		return hosts
	}
	sort.Slice(hosts, func(i, j int) bool {
		switch order {
		case sortOrderAsc:
			return sortByField(hosts, by, i, j)
		case sortOrderDesc:
			return sortByField(hosts, by, j, i)
		}
		return false
	})
	return hosts
}

func sortByField(hosts []Host, field string, i, j int) bool {
	switch field {
	case sortFieldNameLabel:
		return hosts[i].NameLabel < hosts[j].NameLabel
	case sortFieldId:
		return hosts[i].Id < hosts[j].Id
	}
	return false
}
