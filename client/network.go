package client

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
)

type Network struct {
	Automatic       bool     `json:"automatic,omitempty" mapstructure:"automatic,omitempty"`
	Id              string   `json:"id" mapstructure:"id"`
	NameLabel       string   `json:"name_label" mapstructure:"name_label,omitempty"`
	NameDescription string   `json:"name_description" mapstructure:"name_description,omitempty"`
	Bridge          string   `json:"bridge" mapstructure:",omitempty"`
	DefaultIsLocked bool     `json:"defaultIsLocked" mapstructure:"defaultIsLocked,omitempty"`
	PoolId          string   `json:"$poolId" mapstructure:",omitempty"`
	MTU             int      `json:"MTU" mapstructure:",omitempty"`
	PIFs            []string `json:"PIFs" mapstructure:",omitempty"`
	Nbd             bool     `json:"nbd" mapstructure:"nbd,omitempty"`
	InsecureNbd     bool     `json:"insecureNbd" mapstructure:",omitempty"`
}

func (net Network) Compare(obj interface{}) bool {
	otherNet := obj.(Network)
	if net.Id == otherNet.Id {
		return true
	}

	labelsMatch := false
	if net.NameLabel != "" && net.NameLabel == otherNet.NameLabel {
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
		"pool":        netReq.PoolId,
		"name":        netReq.NameLabel,
		"description": netReq.NameDescription,
		"mtu":         netReq.MTU,
		"nbd":         netReq.Nbd,
	}

	err := c.Call("network.create", params, &id)

	if err != nil {
		return nil, err
	}

	if err := c.waitForModifyNetwork(id, netReq.Nbd, time.Minute); err != nil {
		return nil, err
	}
	return c.GetNetwork(Network{Id: id})
}

func (c *Client) waitForModifyNetwork(id string, nbdTarget bool, timeout time.Duration) error {
	if !nbdTarget {
		return nil
	}
	// NBD network creation is eventually consistent so we must poll until
	// the NBD field returns true
	refreshFn := func() (result interface{}, state string, err error) {
		network, err := c.GetNetwork(Network{Id: id})

		if err != nil {
			return network, "", err
		}

		return network, strconv.FormatBool(network.Nbd), nil
	}
	stateConf := &StateChangeConf{
		Pending: []string{"false"},
		Refresh: refreshFn,
		Target:  []string{"true"},
		Timeout: timeout,
	}
	_, err := stateConf.WaitForState()
	return err
}

func (c *Client) UpdateNetwork(netReq Network) (*Network, error) {
	var params map[string]interface{}
	mapstructure.Decode(netReq, &params)
	log.Printf("[DEBUG] params for network.set: %#v", params)

	var success bool
	err := c.Call("network.set", params, &success)
	if err != nil {
		return nil, err
	}

	return c.GetNetwork(netReq)
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

func FindNetworkForTests(poolId string, network *Network) {
	netName, found := os.LookupEnv("XOA_NETWORK")

	if !found {
		fmt.Println("The XOA_NETWORK environment variable must be set")
		os.Exit(-1)
	}

	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		fmt.Printf("failed to create client with error: %v", err)
		os.Exit(-1)
	}

	net, err := c.GetNetwork(Network{
		PoolId:    poolId,
		NameLabel: netName,
	})

	if err != nil {
		fmt.Printf("[ERROR] Failed to get network with error: %v", err)
		os.Exit(1)
	}

	*network = *net
}
