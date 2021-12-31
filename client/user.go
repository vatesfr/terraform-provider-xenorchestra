package client

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
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

func (c *Client) GetAllUsers() ([]User, error) {
	params := map[string]interface{}{
		"dummy": "dummy",
	}
	users := []User{}
	log.Printf("[DEBUG] Calling user.getAll\n")
	err := c.Call("user.getAll", params, &users)

	log.Printf("[DEBUG] Found the following users: %v\n", users)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (c *Client) GetCurrentUser() (*User, error) {
	params := map[string]interface{}{
		"dummy": "dummy",
	}
	user := User{}
	err := c.Call("session.getUser", params, &user)

	log.Printf("[DEBUG] Found the following user: %v with error: %v\n", user, err)
	if err != nil {
		return nil, err
	}
	return &user, err
}

func (c *Client) GetUser(userReq User) (*User, error) {
	users, err := c.GetAllUsers()
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

func RemoveUsersWithPrefix(usernamePrefix string) func(string) error {
	return func(_ string) error {
		c, err := NewClient(GetConfigFromEnv())
		if err != nil {
			return fmt.Errorf("error getting client: %s", err)
		}

		users, err := c.GetAllUsers()

		if err != nil {
			return err
		}

		for _, user := range users {

			if strings.HasPrefix(user.Email, usernamePrefix) {

				log.Printf("[DEBUG] Removing user `%s`\n", user.Email)
				err = c.DeleteUser(user)

				if err != nil {
					log.Printf("failed to remove user `%s` during sweep: %v\n", user.Email, err)
				}
			}
		}
		return nil
	}
}

func CreateUser(user *User) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		fmt.Printf("failed to created client with error: %v", err)
		os.Exit(-1)
	}

	u, err := c.CreateUser(User{
		Email:    user.Email,
		Password: "password",
	})

	if err != nil {
		fmt.Printf("failed to create user for acceptance tests with error: %v", err)
		os.Exit(-1)
	}

	*user = *u
}
