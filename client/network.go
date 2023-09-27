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
	Automatic       bool     `json:"automatic,omitempty"`
	Id              string   `json:"id"`
	NameLabel       string   `json:"name_label"`
	NameDescription string   `json:"name_description"`
	Bridge          string   `json:"bridge"`
	DefaultIsLocked bool     `json:"defaultIsLocked"`
	PoolId          string   `json:"$poolId"`
	MTU             int      `json:"MTU"`
	PIFs            []string `json:"PIFs"`
	Nbd             bool     `json:"nbd"`
	InsecureNbd     bool     `json:"insecureNbd"`
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

type CreateNetworkRequest struct {
	// The first set of members are shared between bonded and non bonded networks
	// These should be kept in sync with the CreateBondedNetworkRequest struct
	// Refactoring these fields to an embedded struct means that the caller must
	// know the embedded structs existance. Since the list is relatively small
	// this was deemed a more appropriate tradeoff to make the API nicer.
	Automatic       bool   `mapstructure:"automatic"`
	DefaultIsLocked bool   `mapstructure:"defaultIsLocked"`
	Pool            string `mapstructure:"pool"`
	Name            string `mapstructure:"name"`
	Description     string `mapstructure:"description,omitempty"`
	Mtu             int    `mapstructure:"mtu,omitempty"`

	Nbd  bool   `mapstructure:"nbd,omitempty"`
	PIF  string `mapstructure:"pif,omitempty"`
	Vlan int    `mapstructure:"vlan,omitempty"`
}

// Nbd and Automatic are eventually consistent. This ensures that waitForModifyNetwork will
// poll until the values are correct.
func (c CreateNetworkRequest) Propagated(obj interface{}) bool {
	otherNet := obj.(Network)

	if otherNet.Automatic == c.Automatic &&
		otherNet.Nbd == c.Nbd {
		return true
	}
	return false
}

type CreateBondedNetworkRequest struct {
	// The first set of members are shared between bonded and non bonded networks
	// These should be kept in sync with the CreateNetworkRequest struct
	Automatic       bool   `mapstructure:"automatic"`
	DefaultIsLocked bool   `mapstructure:"defaultIsLocked"`
	Pool            string `mapstructure:"pool"`
	Name            string `mapstructure:"name"`
	Description     string `mapstructure:"description,omitempty"`
	Mtu             int    `mapstructure:"mtu,omitempty"`

	BondMode string   `mapstructure:"bondMode,omitempty"`
	PIFs     []string `mapstructure:"pifs,omitempty"`
}

func (c CreateBondedNetworkRequest) Propagated(obj interface{}) bool {
	return true
}

type UpdateNetworkRequest struct {
	Id              string  `mapstructure:"id"`
	Automatic       bool    `mapstructure:"automatic"`
	DefaultIsLocked bool    `mapstructure:"defaultIsLocked"`
	NameDescription *string `mapstructure:"name_description,omitempty"`
	NameLabel       *string `mapstructure:"name_label,omitempty"`
	Nbd             bool    `mapstructure:"nbd"`
}

// Nbd and Automatic are eventually consistent. This ensures that waitForModifyNetwork will
// poll until the values are correct.
func (c UpdateNetworkRequest) Propagated(obj interface{}) bool {
	otherNet := obj.(Network)

	if otherNet.Automatic == c.Automatic &&
		otherNet.Nbd == c.Nbd {
		return true
	}
	return false
}

func (c *Client) CreateNetwork(netReq CreateNetworkRequest) (*Network, error) {
	var id string
	var params map[string]interface{}
	mapstructure.Decode(netReq, &params)

	delete(params, "automatic")
	delete(params, "defaultIsLocked")

	log.Printf("[DEBUG] params for network.create: %#v", params)
	err := c.Call("network.create", params, &id)

	if err != nil {
		return nil, err
	}

	// Neither automatic nor defaultIsLocked can be specified in the network.create RPC.
	// Update them afterwards if the user requested it during creation.
	if netReq.Automatic || netReq.DefaultIsLocked {
		_, err = c.UpdateNetwork(UpdateNetworkRequest{
			Id:              id,
			Automatic:       netReq.Automatic,
			DefaultIsLocked: netReq.DefaultIsLocked,
		})
	}

	return c.waitForModifyNetwork(id, netReq, 10*time.Second)
}

func (c *Client) CreateBondedNetwork(netReq CreateBondedNetworkRequest) (*Network, error) {
	var params map[string]interface{}
	mapstructure.Decode(netReq, &params)

	delete(params, "automatic")
	delete(params, "defaultIsLocked")

	log.Printf("[DEBUG] params for network.createBonded: %#v", params)

	var result map[string]interface{}
	err := c.Call("network.createBonded", params, &result)
	if err != nil {
		return nil, err
	}

	id := result["uuid"].(string)
	// Neither automatic nor defaultIsLocked can be specified in the network.create RPC.
	// Update them afterwards if the user requested it during creation.
	if netReq.Automatic || netReq.DefaultIsLocked {
		_, err = c.UpdateNetwork(UpdateNetworkRequest{
			Id:              id,
			Automatic:       netReq.Automatic,
			DefaultIsLocked: netReq.DefaultIsLocked,
		})
	}
	return c.waitForModifyNetwork(id, netReq, 10*time.Second)
}

func (c *Client) waitForModifyNetwork(id string, target RefreshComparison, timeout time.Duration) (*Network, error) {
	refreshFn := func() (result interface{}, state string, err error) {
		network, err := c.GetNetwork(Network{Id: id})

		if err != nil {
			return network, "", err
		}

		equal := strconv.FormatBool(target.Propagated(*network))

		return network, equal, nil
	}
	stateConf := &StateChangeConf{
		Pending: []string{"false"},
		Refresh: refreshFn,
		Target:  []string{"true"},
		Timeout: timeout,
	}
	network, err := stateConf.WaitForState()
	return network.(*Network), err
}

func (c *Client) UpdateNetwork(netReq UpdateNetworkRequest) (*Network, error) {
	var params map[string]interface{}
	mapstructure.Decode(netReq, &params)

	var success bool
	err := c.Call("network.set", params, &success)
	if err != nil {
		return nil, err
	}

	return c.waitForModifyNetwork(netReq.Id, netReq, 10*time.Second)
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
