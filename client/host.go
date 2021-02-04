package client

import (
	"fmt"
	"os"
)

type Host struct {
	Id        string `json:"id"`
	NameLabel string `json:"name_label"`
}

func (h Host) Compare(obj interface{}) bool {
	otherHost := obj.(Host)
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

func FindHostForTests(host *Host) {
	hostName, found := os.LookupEnv("XOA_HOST")

	if !found {
		fmt.Println("The XOA_HOST environment variable must be set")
		os.Exit(-1)
	}
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		fmt.Printf("failed to create client with error: %v", err)
		os.Exit(-1)
	}

	hosts, err := c.GetHostByName(hostName)

	if err != nil {
		fmt.Printf("failed to find a hosts with name: %v with error: %v\n", hostName, err)
		os.Exit(-1)
	}

	if len(hosts) != 1 {
		fmt.Printf("Found %d hosts with name_label %s. Please use a label that is unique so tests are reproducible.\n", len(hosts), hostName)
		os.Exit(-1)
	}

	*host = hosts[0]
}
