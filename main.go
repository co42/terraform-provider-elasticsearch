package main

import (
	"github.com/disaster37/terraform-provider-elasticsearch/es"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: es.Provider,
	})
}
