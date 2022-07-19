package client

import (
	"fmt"
	"log"
	"strings"
)

type CloudConfig struct {
	Name     string `json:"name"`
	Template string `json:"template"`
	Id       string `json:"id"`
}

func (c CloudConfig) Compare(obj interface{}) bool {
	other := obj.(CloudConfig)

	if other.Id == c.Id {
		return true
	}

	if other.Name == c.Name {
		return true
	}

	return false
}

type CloudConfigResponse struct {
	Result []CloudConfig `json:"result"`
}

func (c *Client) GetCloudConfig(id string) (*CloudConfig, error) {
	cloudConfigs, err := c.GetAllCloudConfigs()

	if err != nil {
		return nil, err
	}

	cloudConfig := CloudConfig{Id: id}
	for _, config := range cloudConfigs {
		if cloudConfig.Compare(config) {
			return &config, nil
		}
	}

	// TODO: This should return a NotFound error (see https://github.com/terra-farm/terraform-provider-xenorchestra/issues/118)
	// for more details
	return nil, nil
}

func (c *Client) GetCloudConfigByName(name string) ([]CloudConfig, error) {
	allCloudConfigs, err := c.GetAllCloudConfigs()

	if err != nil {
		return nil, err
	}

	cloudConfigs := []CloudConfig{}
	cloudConfig := CloudConfig{Name: name}
	for _, config := range allCloudConfigs {
		if cloudConfig.Compare(config) {
			cloudConfigs = append(cloudConfigs, config)
		}
	}

	if len(cloudConfigs) == 0 {
		return nil, NotFound{Query: CloudConfig{Name: name}}
	}
	return cloudConfigs, nil
}

func (c *Client) GetAllCloudConfigs() ([]CloudConfig, error) {
	var getAllResp CloudConfigResponse
	params := map[string]interface{}{}
	err := c.Call("cloudConfig.getAll", params, &getAllResp.Result)

	if err != nil {
		return nil, err
	}
	return getAllResp.Result, nil
}

func (c *Client) CreateCloudConfig(name, template string) (*CloudConfig, error) {
	params := map[string]interface{}{
		"name":     name,
		"template": template,
	}
	// Xen Orchestra versions >= 5.98.0 changed this return value to an object
	// when older versions returned bool. This needs to be an interface
	// type in order to be backwards compatible while fixing this bug. See
	// GitHub issue 204 for more details.
	var resp interface{}
	err := c.Call("cloudConfig.create", params, &resp)

	if err != nil {
		return nil, err
	}

	// Since the Id isn't returned in the reponse loop over all cloud configs
	// and find the one we just created
	cloudConfigs, err := c.GetAllCloudConfigs()

	if err != nil {
		return nil, err
	}

	var found CloudConfig
	for _, config := range cloudConfigs {
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
	err := c.Call("cloudConfig.delete", params, &resp)

	if err != nil {
		return err
	}

	return nil
}

func RemoveCloudConfigsWithPrefix(cloudConfigPrefix string) func(string) error {
	return func(_ string) error {
		c, err := NewClient(GetConfigFromEnv())
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		cloudConfigs, err := c.GetAllCloudConfigs()

		if err != nil {
			return err
		}

		for _, cloudConfig := range cloudConfigs {

			if strings.HasPrefix(cloudConfig.Name, cloudConfigPrefix) {

				log.Printf("[DEBUG] Removing cloud config `%s`\n", cloudConfig.Name)
				err = c.DeleteCloudConfig(cloudConfig.Id)

				if err != nil {
					log.Printf("failed to remove cloud config `%s` during sweep: %v\n", cloudConfig.Name, err)
				}
			}
		}
		return nil
	}
}
