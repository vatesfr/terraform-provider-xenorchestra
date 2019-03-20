package main

import (
	"fmt"

	"github.com/ddelnano/terraform-provider-xenorchestra/client"
)

func main() {
	c, err := client.NewClient()

	if err != nil {
		fmt.Errorf("Failed to create a client: %v", err)
	}

	tmpl, err := c.GetTemplate("Ubuntu Bionic Beaver 18.04")

	if err != nil {
		fmt.Printf("error when getting template: %v\n", err)
	}

	fmt.Printf("Template %#v", tmpl)

	// 	vm, err := c.GetVm("77c6637c-fa3d-0a46-717e-296208c40169")

	// 	if err != nil {
	// 		fmt.Errorf("failed to get vm %v", err)
	// 	}
	// 	fmt.Printf("vm: %#v", vm)

	// 	vm, err = c.CreateVm(
	// 		"client_test",
	// 		"client_test description",
	// 		"2dd0373e-0ed5-7413-a57f-1958d03b698c",
	// 		"ad9d4c6b-3aee-4143-bcf0-bee546664093",
	// 		1,
	// 		1073733632)

	// 	if err != nil {
	// 		fmt.Printf("faled to create vm: %v", err)
	// 	}

	// 	fmt.Printf("Created vm %#v", vm)

	vm, err := c.GetVm("3c724315-ad91-31c5-a153-6ae4f4fc825b")

	if err != nil {
		fmt.Printf("failed to retrieve vm %#v", err)
	}

	fmt.Printf("retrieved vm %#v", vm)
	// cloudConfig, err := c.CreateCloudConfig("testing", "new client")

	// if err != nil {
	// 	fmt.Errorf("failed to create a cloud config: %v", err)
	// }

	// fmt.Printf("cloud config: %v\n", cloudConfig)

	// cloudConfig, err = c.GetCloudConfig(cloudConfig.Id)

	// fmt.Printf("cloud config: %v\n", cloudConfig)
	// vm, err := c.CreateVm(
	// 	"testing",
	// 	"description",
	// 	"2dd0373e-0ed5-7413-a57f-1958d03b698c",
	// 	"06e49231-34bc-40d8-ba8a-d2bef5161177",
	// 	1,
	// 	10737336,
	// )

	// if err != nil {
	// 	fmt.Errorf("Failed to create a vm: %v", err)
	// }

	// fmt.Printf("vm %v", vm)
}
