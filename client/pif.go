package client

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type PIF struct {
	Device   string
	Host     string
	Network  string
	Id       string
	Uuid     string
	PoolId   string
	Attached bool
}

func (c *Client) GetPIFByDevice(dev string) (PIF, error) {
	var pif PIF
	params := map[string]interface{}{
		"type": "PIF",
	}
	var objsRes struct {
		PIFs map[string]interface{} `json:"-"`
	}
	ctx, _ := context.WithTimeout(context.Background(), 100*time.Second)
	err := c.rpc.Call(ctx, "xo.getAllObjects", params, &objsRes.PIFs)

	if err != nil {
		return pif, err
	}

	found := false
	for _, result := range objsRes.PIFs {

		v, ok := result.(map[string]interface{})
		if !ok {
			return pif, errors.New("Could not coerce interface{} into map")
		}

		if v["type"].(string) != "PIF" {
			continue
		}

		pifDev, ok := v["device"].(string)

		if !ok {
			return pif, errors.New(fmt.Sprintf("type assertion for device failed on PIF: %v", v))
		}

		if pifDev == dev {
			found = true
			id := v["id"].(string)
			attached := v["attached"].(bool)
			uuid := v["uuid"].(string)
			poolId := v["$poolId"].(string)
			host := v["$host"].(string)
			network := v["$network"].(string)
			pif = PIF{
				Device:   pifDev,
				Host:     host,
				Network:  network,
				Id:       id,
				Uuid:     uuid,
				PoolId:   poolId,
				Attached: attached,
			}
		}
	}

	if !found {
		return pif, NotFound{Type: "PIF"}
	}

	return pif, nil
}
