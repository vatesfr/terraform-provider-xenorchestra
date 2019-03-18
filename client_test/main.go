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

	vm, err := c.CreateVm(
		"testing",
		"description",
		"2dd0373e-0ed5-7413-a57f-1958d03b698c",
		"06e49231-34bc-40d8-ba8a-d2bef5161177",
		1,
		10737336,
	)

	if err != nil {
		fmt.Errorf("Failed to create a vm: %v", err)
	}

	fmt.Printf("vm %v", vm)
}
