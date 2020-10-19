package client

import (
	"errors"
	"fmt"
	"log"
)

func (c *Client) AddTag(id, tag string) error {
	var success bool
	params := map[string]interface{}{
		"id":  id,
		"tag": tag,
	}
	err := c.Call("tag.add", params, &success)

	if err != nil {
		return err
	}
	return nil
}

func (c *Client) RemoveTag(id, tag string) error {
	var success bool
	params := map[string]interface{}{
		"id":  id,
		"tag": tag,
	}
	err := c.Call("tag.remove", params, &success)

	if err != nil {
		return err
	}
	return nil
}

func (c *Client) GetObjectsWithTags(tags []string) ([]string, error) {
	var objsRes struct {
		Objects map[string]interface{} `json:"-"`
	}
	params := map[string]interface{}{
		"filter": map[string][]string{
			"tags": tags,
		},
	}
	c.Call("xo.getAllObjects", params, &objsRes.Objects)
	log.Printf("[DEBUG] Found objects with tags `%s`: %v\n", tags, objsRes)

	t := []string{}
	for _, resObject := range objsRes.Objects {
		obj, ok := resObject.(map[string]interface{})

		if !ok {
			return t, errors.New("Could not coerce interface{} into map")
		}

		id := obj["id"].(string)
		t = append(t, id)
	}
	return t, nil
}

func RemoveTagFromAllObjects(tag string) func(string) error {
	return func(_ string) error {
		c, err := NewClient(GetConfigFromEnv())
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		objects, err := c.GetObjectsWithTags([]string{tag})

		if err != nil {
			return err
		}

		for _, object := range objects {
			log.Printf("[DEBUG] Remove tag `%s` on object `%s`\n", tag, object)
			err = c.RemoveTag(object, tag)

			if err != nil {
				log.Printf("error remove tag `%s` during sweep: %v", tag, err)
			}
		}
		return nil
	}
}
