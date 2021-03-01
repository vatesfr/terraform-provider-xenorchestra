package client

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
)

type Host struct {
	Id        string        `json:"id"`
	NameLabel string        `json:"name_label"`
	Tags      []interface{} `json:"tags,omitempty"`
	Pool      string        `json:"$pool"`
}

func (h Host) Compare(obj interface{}) bool {
	otherHost := obj.(Host)
	if otherHost.Id == h.Id {
		return true
	}
	if h.Pool == otherHost.Pool {
		return true
	}
	if otherHost.NameLabel != "" && h.NameLabel == otherHost.NameLabel {
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

func (c *Client) GetHostsByPoolName(host Host) (hosts []map[string]interface{}, err error) {
	hosts = make([]map[string]interface{}, 0)
	obj, err := c.FindFromGetAllObjects(host)

	if err != nil {
		return
	}
	slice := obj.([]Host)
	for _, v := range slice {
		hosts = append(hosts, convertToMap(v))
	}
	return hosts, nil
}

func convertToMap(st interface{}) map[string]interface{} {
	reqRules := make(map[string]interface{})
	v := reflect.ValueOf(st)
	t := reflect.TypeOf(st)
	for i := 0; i < v.NumField(); i++ {
		key := strings.ToLower(t.Field(i).Name)
		typ := v.FieldByName(t.Field(i).Name).Kind().String()
		structTag := t.Field(i).Tag.Get("json")
		jsonName := strings.TrimSpace(strings.Split(structTag, ",")[0])
		value := v.FieldByName(t.Field(i).Name)
		if jsonName != "" && jsonName != "-" {
			key = jsonName
		}
		if typ == "string" {
			if !(value.String() == "" && strings.Contains(structTag, "omitempty")) {
				fmt.Println(key, value)
				fmt.Println(key, value.String())
				reqRules[key] = value.String()
			}
		} else if typ == "slice" {
			if value.Len() > 0 {
				reqRules[key] = fmt.Sprintf("%s", value)
			}
		} else {
			reqRules[key] = value.String()
		}
	}
	return reqRules
}
