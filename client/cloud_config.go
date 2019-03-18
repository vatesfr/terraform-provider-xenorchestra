package client

type CloudConfig struct {
	Name     string `json:"name"`
	Template string `json:"template"`
	Id       string `json:"id"`
}

type CloudConfigResponse struct {
	Result []CloudConfig `json:"result"`
}

func (c *Client) GetCloudConfig(id string) (*CloudConfig, error) {
	var params map[string]interface{}
	var getAllResp CloudConfigResponse
	err := c.rpc.Call("cloudConfig.getAll", params, &getAllResp.Result)

	if err != nil {
		return nil, err
	}

	var configResult CloudConfig
	found := false
	for _, config := range getAllResp.Result {
		if config.Id == id {
			configResult = config
			found = true
		}
	}

	if !found {
		return nil, nil
	}
	return &configResult, nil
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

	// Since the Id isn't returned in the reponse loop over all cloud configs
	// and find the one we just created
	var getAllResp CloudConfigResponse
	err = c.rpc.Call("cloudConfig.getAll", params, &getAllResp.Result)

	if err != nil {
		return nil, err
	}

	var found CloudConfig
	for _, config := range getAllResp.Result {
		if config.Name == name && config.Template == template {
			found = config
		}
	}
	return &found, nil
}

func (c *Client) DeleteCloudConfig(id string) error {
	params := map[string]interface{}{
		"id": id,
	}
	var resp bool
	err := c.rpc.Call("cloudConfig.delete", params, &resp)

	if err != nil {
		return err
	}

	return nil
}
