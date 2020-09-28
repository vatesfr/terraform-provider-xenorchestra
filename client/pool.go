package client

import (
	"fmt"
	"os"
)

type Pool struct {
	Id          string
	NameLabel   string
	Description string
	Cpus        struct {
		Cores   int64
		Sockets int64
	}
}

func (p Pool) Compare(obj map[string]interface{}) bool {
	nameLabel := obj["name_label"].(string)

	if nameLabel != p.NameLabel {
		return false
	}
	return true
}

func (p Pool) New(obj map[string]interface{}) XoObject {
	nameLabel := obj["name_label"].(string)
	id := obj["id"].(string)
	description := obj["name_description"].(string)
	cpus := obj["cpus"].(map[string]interface{})
	cores := cpus["cores"].(float64)
	sockets := cpus["sockets"].(float64)
	return Pool{
		Id:          id,
		NameLabel:   nameLabel,
		Description: description,
		Cpus: struct {
			Cores   int64
			Sockets int64
		}{
			Cores:   int64(cores),
			Sockets: int64(sockets),
		},
	}
}

func (c *Client) GetPoolByName(name string) (Pool, error) {
	obj, err := c.FindFromGetAllObjects(Pool{NameLabel: name})
	pool := obj.(Pool)

	if err != nil {
		return pool, err
	}

	return pool, nil
}

func FindPoolForTests(pool *Pool) {
	poolName, found := os.LookupEnv("XOA_POOL")

	if !found {
		fmt.Println("The XOA_POOL environment variable must be set")
		os.Exit(-1)
	}
	c, _ := NewClient(GetConfigFromEnv())
	var err error
	*pool, err = c.GetPoolByName(poolName)

	if err != nil {
		fmt.Printf("failed to find a pool with name: %v with error: %v\n", poolName, err)
		os.Exit(-1)
	}
}
