package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
)

type Network struct {
	Id        string `json:"id"`
	NameLabel string `json:"name_label"`
	Bridge    string
	PoolId    string
}

func (net Network) Compare(obj map[string]interface{}) bool {
	id := obj["id"].(string)
	nameLabel := obj["name_label"].(string)
	if net.Id == id {
		return true
	}

	if net.NameLabel == nameLabel {
		return true
	}

	return false
}

func (net Network) New(obj map[string]interface{}) XoObject {
	id := obj["id"].(string)
	poolId := obj["$poolId"].(string)
	nameLabel := obj["name_label"].(string)
	return Network{
		Id:        id,
		PoolId:    poolId,
		NameLabel: nameLabel,
	}
}

func (c *Client) CreateNetwork(netReq Network) (*Network, error) {
	var id string
	params := map[string]interface{}{
		"pool": netReq.PoolId,
		"name": netReq.NameLabel,
	}

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := c.Call(ctx, "network.create", params, &id)

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

	if _, ok := obj.([]interface{}); ok {
		return nil, errors.New("Your query returned more than one result. Use `pool_id` or other attributes to filter the result down to a single network")
	}

	net := obj.(Network)
	return &net, nil
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

	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	err := c.Call(ctx, "network.delete", params, &success)

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
