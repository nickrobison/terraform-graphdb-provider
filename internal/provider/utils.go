package provider

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func ProviderDataError(data any, diags *diag.Diagnostics) {
	diags.AddError(
		"Unexpected Resource Configure Type",
		fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", data))
}
