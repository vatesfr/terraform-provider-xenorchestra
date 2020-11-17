package client

import (
	"errors"
	"fmt"
	"log"
	"strings"
)

type Network struct {
	Id        string `json:"id"`
	NameLabel string `json:"name_label"`
	Bridge    string `json:"bridge"`
	PoolId    string `json:"$poolId"`
}

func (net Network) Compare(obj interface{}) bool {
	otherNet := obj.(Network)
	if net.Id == otherNet.Id {
		return true
	}

	labelsMatch := false
	if net.NameLabel == otherNet.NameLabel {
		labelsMatch = true
	}

	if net.PoolId == "" && labelsMatch {
		return true
	} else if net.PoolId == otherNet.PoolId && labelsMatch {
		return true
	}

	return false
}

func (c *Client) CreateNetwork(netReq Network) (*Network, error) {
	var id string
	params := map[string]interface{}{
		"pool": netReq.PoolId,
		"name": netReq.NameLabel,
	}

	err := c.Call("network.create", params, &id)

	if err != nil {
		return nil, err
	}
	return c.GetNetwork(Network{Id: id})
}

func (c *Client) GetNetwork(netReq Network) (*Network, error) {
	obj, err := c.FindFromGetAllObjects(netReq)

	if err != nil {
		return nil, err
	}

	nets := obj.([]Network)

	if len(nets) > 1 {
		return nil, errors.New(fmt.Sprintf("Your query returned more than one result: %+v. Use `pool_id` or other fields to filter the result down to a single network", nets))
	}

	return &nets[0], nil
}

func (c *Client) GetNetworks() ([]Network, error) {
	var response map[string]Network
	err := c.GetAllObjectsOfType(Network{}, &response)

	nets := make([]Network, 0, len(response))
	for _, net := range response {
		nets = append(nets, net)
	}
	return nets, err
}

func (c *Client) DeleteNetwork(id string) error {
	var success bool
	params := map[string]interface{}{
		"id": id,
	}

	err := c.Call("network.delete", params, &success)

	return err
}

func RemoveNetworksWithNamePrefix(prefix string) func(string) error {
	return func(_ string) error {
		c, err := NewClient(GetConfigFromEnv())
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		nets, err := c.GetNetworks()

		if err != nil {
			return err
		}

		for _, net := range nets {
			if strings.HasPrefix(net.NameLabel, prefix) {
				log.Printf("[DEBUG] Deleting network: %v\n", net)
				err = c.DeleteNetwork(net.Id)

				if err != nil {
					log.Printf("error destroying network `%s` during sweep: %v", net.NameLabel, err)
				}
			}
		}
		return nil
	}
}
