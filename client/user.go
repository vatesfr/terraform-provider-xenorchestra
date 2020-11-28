package client

import (
	"errors"
	"log"
)

type User struct {
	Id          string
	Email       string
	Password    string `json:"-"`
	Groups      []string
	Permission  string
	Preferences Preferences
}

type Preferences struct {
	SshKeys []SshKey `json:"sshKeys,omitempty"`
}

type SshKey struct {
	Title string
	Key   string
}

func (user User) Compare(obj interface{}) bool {
	other := obj.(User)

	if user.Id == other.Id {
		return true
	}

	if user.Email == other.Email {
		return true
	}

	return false
}

func (c *Client) CreateUser(user User) (*User, error) {
	var id string
	params := map[string]interface{}{
		"email":    user.Email,
		"password": user.Password,
	}
	err := c.Call("user.create", params, &id)

	if err != nil {
		return nil, err
	}

	return c.GetUser(User{Id: id})
}

func (c *Client) GetUser(userReq User) (*User, error) {
	params := map[string]interface{}{
		"dummy": "dummy",
	}
	users := []User{}
	err := c.Call("user.getAll", params, &users)

	log.Printf("[DEBUG] Found the following users: %v\n", users)
	if err != nil {
		return nil, err
	}

	var foundUser User
	for _, user := range users {
		if user.Compare(userReq) {
			foundUser = user
		}
	}

	if foundUser.Id == "" {
		return nil, NotFound{Query: userReq}
	}

	return &foundUser, nil
}

func (c *Client) DeleteUser(user User) error {
	var success bool
	params := map[string]interface{}{
		"id": user.Id,
	}
	err := c.Call("user.delete", params, &success)

	if err != nil {
		return err
	}

	if !success {
		return errors.New("failed to delete user")
	}
	return nil
}
