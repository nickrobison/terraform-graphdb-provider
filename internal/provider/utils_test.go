package provider

import (
	"testing"
)

func TestRoleConverstion(t *testing.T) {
	role := "repo-manager"
	want := "ROLE_REPO_MANAGER"
	auth := roleToAuthority(role)
	if auth != want {
		t.Fatalf("Failed to convert role. Wanted: %s. Got %s", want, auth)
	}
}

func TestAuthorityConversionSuccess(t *testing.T) {
	auth := "ROLE_REPO_MANAGER"
	want := "repo-manager"
	role, err := authorityToRole(auth)
	if err != nil {
		t.Fatal(err)
	}
	if role != want {
		t.Fatalf("Failed to convert authority. Wanted %s. Got %s", want, role)
	}
}

func TestAuthorityConversionFailure(t *testing.T) {
	auth := "ROLEZ"
	_, err := authorityToRole(auth)
	if err.Error() != "Unsupported authority [ROLEZ]" {
		t.Fatalf("Should have failed to convert invalid authority %s. Got error: %s", auth, err.Error())
	}
}
