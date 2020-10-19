package client

import (
	"fmt"
	"os"
)

type Pool struct {
	Id          string  `json:"id"`
	NameLabel   string  `json:"name_label"`
	Description string  `json:"name_description"`
	Cpus        CpuInfo `json:"cpus"`
	DefaultSR   string  `json:"default_SR"`
	Master      string  `json:"master"`
}

type CpuInfo struct {
	Cores   int64 `json:"cores,float64`
	Sockets int64 `json:"sockets,float64`
}

func (p Pool) Compare(obj interface{}) bool {
	otherPool := obj.(Pool)

	if otherPool.NameLabel != p.NameLabel {
		return false
	}
	return true
}

func (c *Client) GetPoolByName(name string) (pools []Pool, err error) {
	obj, err := c.FindFromGetAllObjects(Pool{NameLabel: name})
	if err != nil {
		return
	}
	pools = obj.([]Pool)

	return pools, nil
}

func FindPoolForTests(pool *Pool) {
	poolName, found := os.LookupEnv("XOA_POOL")

	if !found {
		fmt.Println("The XOA_POOL environment variable must be set")
		os.Exit(-1)
	}
	c, err := NewClient(GetConfigFromEnv())
	if err != nil {
		fmt.Printf("failed to create client with error: %v", err)
		os.Exit(-1)
	}

	pools, err := c.GetPoolByName(poolName)

	if err != nil {
		fmt.Printf("failed to find a pool with name: %v with error: %v\n", poolName, err)
		os.Exit(-1)
	}

	if len(pools) != 1 {
		fmt.Printf("Found %d pools with name_label %s. Please use a label that is unique so tests are reproducible.\n", len(pools), poolName)
		os.Exit(-1)
	}

	*pool = pools[0]
}
