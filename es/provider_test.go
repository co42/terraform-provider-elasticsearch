package es

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
)

var testAccProviders map[string]*schema.Provider
var testAccProvider *schema.Provider

func init() {

	// Init logger
	logrus.SetFormatter(&easy.Formatter{
		LogFormat: "[%lvl%] %msg%\n",
	})
	//logrus.SetLevel(logrus.DebugLevel)

	// Init provider
	testAccProvider = Provider()
	configureFunc := testAccProvider.ConfigureFunc
	testAccProvider.ConfigureFunc = func(d *schema.ResourceData) (interface{}, error) {
		return configureFunc(d)
	}
	testAccProviders = map[string]*schema.Provider{
		"elasticsearch": testAccProvider,
	}

}

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ *schema.Provider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("ELASTICSEARCH_URLS"); v == "" {
		t.Fatal("ELASTICSEARCH_URLS must be set for acceptance tests")
	}

}
