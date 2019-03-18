package client

import "fmt"

type allObjectResponse struct {
	Objects map[string]Vm `json:"-"`
}

type CPUs struct {
	Number int
}

type MemoryObject struct {
	Dynamic []int `json:"dynamic"`
	Static  []int `json:"static"`
	Size    int   `json:"size"`
}

type Vm struct {
	Type               string       `json:"type,omitempty"`
	Id                 string       `json:"id,omitempty"`
	Name               string       `json:"name,omitempty"`
	NameDescription    string       `json:"name_description"`
	NameLabel          string       `json:"name_label"`
	CPUs               CPUs         `json:"CPUs"`
	Memory             MemoryObject `json:"memory"`
	PowerState         string       `json:"power_state"`
	VIFs               []string     `json:"VIFs"`
	VirtualizationMode string       `json:"virtualizationMode"`
	PoolId             string       `json:"$poolId"`
	Template           string       `json:"template"`
	CloudConfig        string       `json:"cloudConfig"`
}

func (c *Client) CreateVm(name_label, name_description, template, cloudConfig string, cpus, memoryMax int) (*Vm, error) {
	params := map[string]interface{}{
		"bootAfterCreate":  true,
		"name_label":       name_label,
		"name_description": name_description,
		"template":         template,
		"cloudConfig":      cloudConfig,
		"coreOs":           false,
		"cpuCap":           nil,
		"cpuWeight":        nil,
		"CPUs":             cpus,
		"memoryMax":        memoryMax,
		"existingDisks": map[string]interface{}{
			"0": map[string]interface{}{
				"$SR":              "7f469400-4a2b-5624-cf62-61e522e50ea1",
				"name_description": "Created by XO",
				"name_label":       "Ubuntu Bionic Beaver 18.04_imavo",
				"size":             32212254720,
			},
			"1": map[string]interface{}{
				"$SR":              "7f469400-4a2b-5624-cf62-61e522e50ea1",
				"name_description": "",
				"name_label":       "XO CloudConfigDrive",
				"size":             10485760,
			},
			"2": map[string]interface{}{
				"$SR":              "7f469400-4a2b-5624-cf62-61e522e50ea1",
				"name_description": "",
				"name_label":       "XO CloudConfigDrive",
				"size":             10485760,
			},
		},
		"VIFs": []interface{}{
			map[string]string{
				"network": "d225cf00-36f8-e6d6-6a29-02636d4de56b",
			},
		},
	}
	var reply struct {
		Params map[string]interface{} `json:"-"`
	}
	err := c.rpc.Call("vm.create", params, &reply.Params)
	fmt.Printf("vm.create reply: %v", reply.Params)

	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *Client) DeleteVm(id string) error {
	params := map[string]interface{}{
		"id": id,
	}
	var reply bool
	err := c.rpc.Call("vm.delete", params, &reply)

	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetVm(id string) (*Vm, error) {
	params := map[string]interface{}{
		"type": "VM",
	}
	var objsRes allObjectResponse
	err := c.rpc.Call("xo.getAllObjects", params, &objsRes.Objects)

	if err != nil {
		return nil, err
	}

	vm, ok := objsRes.Objects[id]
	if !ok {
		return nil, fmt.Errorf("Could not find Vm with id: %s", id)
	}
	return &vm, nil
}
