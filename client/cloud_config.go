package client

import "fmt"

type CloudConfig struct {
	Name     string `json:"name"`
	Template string `json:"template"`
	Id       string `json:"id"`
}

func (c *Client) CreateCloudConfig(name, template string) (*CloudConfig, error) {
	params := map[string]interface{}{
		"name":     name,
		"template": template,
	}
	var resp bool
	err := c.rpc.Call("cloudConfig.create", params, &resp)

	if err != nil {
		return nil, err
	}

	config := &CloudConfig{
		Name:     name,
		Template: template,
	}

	// Since the Id isn't returned in the reponse loop over all cloud configs
	// and find the one we just created

	var getAllResp CloudConfigResponse
	err = c.rpc.Call("cloudConfig.getAll", nil, &getAllResp)

	fmt.Printf("get allresponse %v error: %v", getAllResp, err)
	if err != nil {
		return nil, err
	}
	return config, nil
}
