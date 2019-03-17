package main

import (
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
)

func main() {
	c, err := client.NewClient("192.168.88.86", "admin@admin.net", "admin")

	if err != nil {
		fmt.Errorf("Failed to create a client: %v", err)
	}

	_, err = c.CreateCloudConfig("testing2", "cloud-init")

	if err != nil {
		fmt.Errorf("failed to create cloud config: %v", err)
	}

	// fmt.Println(fmt.Sprintf("name: %s template: %s", config.Name, config.Template))
}
