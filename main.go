/*
Package main create Elasticsearch provider for Terraform

Read the doc to use it: https://github.com/disaster37/terraform-provider-elasticsearch/tree/7.x
*/
package main

import (
	"context"
	"flag"
	"os"

	"github.com/disaster37/terraform-provider-elasticsearch/v7/es"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	log "github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

func init() {

	log.SetOutput(os.Stderr)
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&easy.Formatter{
		LogFormat: "[%lvl%] %msg%",
	})

}

func main() {

	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{ProviderFunc: es.Provider}

	if debugMode {
		err := plugin.Debug(context.Background(), "registry.terraform.io/disaster37/elasticsearch", opts)

		if err != nil {
			log.Fatal(err.Error())
		}

		return
	}

	plugin.Serve(opts)

}
