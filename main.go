/*
Package main create Elasticsearch provider for Terraform

Read the doc to use it: https://github.com/disaster37/terraform-provider-elasticsearch/tree/7.x
*/
package main

import (
	"github.com/disaster37/terraform-provider-elasticsearch/v7/es"
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: es.Provider,
	})
}
