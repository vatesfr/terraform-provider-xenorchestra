package client

import (
	"testing"
)

func TestCreateUserAndDeleteUser(t *testing.T) {
	// c, err := NewClient(GetConfigFromEnv())

	// expectedAcl := Acl{
	// 	Subject: fmt.Sprintf("%s-%s", integrationTestPrefix, "acl"),
	// 	Action:  "viewer",
	// 	Object:  accVm.Id,
	// }

	// if err != nil {
	// 	t.Fatalf("failed to create client with error: %v", err)
	// }

	// acl, err := c.CreateAcl(expectedAcl)

	// if err != nil {
	// 	t.Fatalf("failed to create acl with error: %v", err)
	// }

	// if acl == nil {
	// 	t.Fatalf("expected to receive non-nil Acl")
	// }

	// if acl.Id == "" {
	// 	t.Errorf("expected Acl to have a non-empty Id")
	// }

	// if acl.Subject != expectedAcl.Subject {
	// 	t.Errorf("expected acl's subject `%s` to match `%s`", acl.Subject, expectedAcl.Subject)
	// }

	// err = c.DeleteAcl(expectedAcl)

	// if err != nil {
	// 	t.Errorf("failed to delete acl with error: %v", err)
	// }
}

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
