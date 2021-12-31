package client

import (
	"fmt"
	"testing"
)

func TestGetUser(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	expectedUser := User{
		Email:    "ddelnano",
		Password: "password",
	}

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	user, err := c.CreateUser(expectedUser)
	defer c.DeleteUser(*user)

	if err != nil {
		t.Fatalf("failed to create user with error: %v", err)
	}

	if user == nil {
		t.Fatalf("expected to receive non-nil user")
	}

	if user.Id == "" {
		t.Errorf("expected user to have a non-empty Id")
	}

	_, err = c.GetUser(User{Id: user.Id})

	if err != nil {
		t.Errorf("failed to find user by id `%s` with error: %v", user.Id, err)
	}
}

func TestGetCurrentUser(t *testing.T) {
	c, err := NewClient(GetConfigFromEnv())

	if err != nil {
		t.Fatalf("failed to create client with error: %v", err)
	}

	user, err := c.GetCurrentUser()

	if err != nil {
		t.Fatalf("failed to retrieve the current user with error: %v", err)
	}

	fmt.Printf("Found user: %v", user)
}
