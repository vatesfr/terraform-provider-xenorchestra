package client

import (
	"errors"
	"fmt"
	"os"
)

type PIF struct {
	Device       string `json:"device"`
	Host         string `json:"$host"`
	Network      string `json:"$network"`
	Id           string `json:"id"`
	Uuid         string `json:"uuid"`
	PoolId       string `json:"$poolId"`
	Attached     bool   `json:"attached"`
	Vlan         int    `json:"vlan"`
	IsBondMaster bool   `json:"isBondMaster,omitempty"`
	IsBondSlave  bool   `json:"isBondSlave,omitempty"`
}

func (p PIF) Compare(obj interface{}) bool {
	otherPif := obj.(PIF)

	if p.Id != "" {
		if otherPif.Id == p.Id {
			return true
		} else {
			return false
		}
	}
	hostIdExists := p.Host != ""
	if hostIdExists && p.Host != otherPif.Host {
		return false
	}

	networkIdExists := p.Network != ""
	if networkIdExists && p.Network != otherPif.Network {
		return false
	}

	if p.Vlan == otherPif.Vlan && (p.Device == "" || (p.Device == otherPif.Device)) {
		return true
	}
	return false
}

func (c *Client) GetPIFByDevice(dev string, vlan int) ([]PIF, error) {
	obj, err := c.FindFromGetAllObjects(PIF{Device: dev, Vlan: vlan})

	if err != nil {
		return []PIF{}, err
	}
	pifs, ok := obj.([]PIF)

	if !ok {
		return pifs, errors.New("failed to coerce response into PIF slice")
	}

	return pifs, nil
}

func (c *Client) GetPIF(pifReq PIF) (pifs []PIF, err error) {
	obj, err := c.FindFromGetAllObjects(pifReq)

	if err != nil {
		return
	}
	pifs, ok := obj.([]PIF)

	if !ok {
		return pifs, errors.New("failed to coerce response into PIF slice")
	}

	return pifs, nil
}

func FindPIFForTests(pif *PIF) {
	pifId, found := os.LookupEnv("XOA_PIF")

	if !found {
		fmt.Println("The XOA_PIF environment variable must be set to run the network resource tests")
		return
	}

	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		fmt.Printf("failed to create client with error: %v", err)
		os.Exit(-1)
	}

	pifs, err := c.GetPIF(PIF{Id: pifId})

	if err != nil {
		fmt.Printf("[ERROR] Failed to get pif with error: %v", err)
		os.Exit(1)
	}

	if len(pifs) != 1 {
		fmt.Printf("[ERROR] expected to find a single pif. Found %d PIFs instead: %v", len(pifs), pifs)
		os.Exit(1)
	}

	*pif = pifs[0]
}
