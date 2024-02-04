package provider

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func ProviderDataError(data any, diags *diag.Diagnostics) {
	diags.AddError(
		"Unexpected Resource Configure Type",
		fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", data))
}

func roleToAuthority(role string) string {
	return fmt.Sprintf("ROLE_%s", strings.Replace(strings.ToUpper(role), "-", "_", 1))
}

func authorityToRole(authority string) (string, error) {
	split := strings.SplitN(authority, "_", 2)
	if len(split) < 2 {
		return "", fmt.Errorf("Unsupported authority %s", split)
	}
	return strings.ToLower(strings.Replace(split[1], "_", "-", 1)), nil
}
