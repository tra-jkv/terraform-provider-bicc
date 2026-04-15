package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/tra-jkv/terraform-provider-bicc/bicc"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: bicc.Provider,
	})
}
