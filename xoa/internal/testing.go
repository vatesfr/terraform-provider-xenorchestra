package internal

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

/*
 * This file contains code that was borrowed from the terraform-provider-aws repo (https://github.com/terraform-providers/terraform-provider-aws). It is provided as is (no modifications) less the code that I did not use.
 *
 * It is covered by the Mozilla Public License Version 2.0.
 *
 * See https://github.com/terraform-providers/terraform-provider-aws/blob/master/LICENSE for copyright and licensing information.
 */

const (
	sentinelIndex = "*"
)

// instanceState returns the primary instance state for the given
// resource name in the root module.
func instanceState(s *terraform.State, name string) (*terraform.InstanceState, error) {
	ms := s.RootModule()
	rs, ok := ms.Resources[name]
	if !ok {
		return nil, fmt.Errorf("Not found: %s in %s", name, ms.Path)
	}

	is := rs.Primary
	if is == nil {
		return nil, fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
	}

	return is, nil
}

// TestCheckTypeSetElemAttrPair is a TestCheckFunc that verifies a pair of name/key
// combinations are equal where the first uses the sentinel value to index into a
// TypeSet.
//
// E.g., tfawsresource.TestCheckTypeSetElemAttrPair("aws_autoscaling_group.bar", "availability_zones.*", "data.aws_availability_zones.available", "names.0")
func TestCheckTypeSetElemAttrPair(nameFirst, keyFirst, nameSecond, keySecond string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		isFirst, err := instanceState(s, nameFirst)
		if err != nil {
			return err
		}

		isSecond, err := instanceState(s, nameSecond)
		if err != nil {
			return err
		}

		vSecond, okSecond := isSecond.Attributes[keySecond]
		if !okSecond {
			return fmt.Errorf("%s: Attribute %q not set, cannot be checked against TypeSet", nameSecond, keySecond)
		}

		return testCheckTypeSetElem(isFirst, keyFirst, vSecond)
	}
}

func testCheckTypeSetElem(is *terraform.InstanceState, attr, value string) error {
	attrParts := strings.Split(attr, ".")
	if attrParts[len(attrParts)-1] != sentinelIndex {
		return fmt.Errorf("%q does not end with the special value %q", attr, sentinelIndex)
	}
	for stateKey, stateValue := range is.Attributes {
		if stateValue == value {
			stateKeyParts := strings.Split(stateKey, ".")
			if len(stateKeyParts) == len(attrParts) {
				for i := range attrParts {
					if attrParts[i] != stateKeyParts[i] && attrParts[i] != sentinelIndex {
						break
					}
					if i == len(attrParts)-1 {
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("no TypeSet element %q, with value %q in state: %#v", attr, value, is.Attributes)
}
