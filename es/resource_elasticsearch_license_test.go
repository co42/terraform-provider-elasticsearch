package es

import (
	"context"
	"fmt"
	"testing"

	elastic7 "github.com/elastic/go-elasticsearch/v7"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchLicense(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchLicenseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchLicense,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchLicenseExists("elasticsearch_license.test"),
				),
			},
		},
	})
}

func testCheckElasticsearchLicenseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No license ID is set")
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			res, err := client.API.License.Get(
				client.API.License.Get.WithContext(context.Background()),
				client.API.License.Get.WithPretty(),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				return errors.Errorf("Error when get license: %s", res.String())
			}
		default:
			return errors.New("License is only supported by the elastic library >= v6!")
		}

		return nil
	}
}

func testCheckElasticsearchLicenseDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_license" {
			continue
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			res, err := client.API.License.Get(
				client.API.License.Get.WithContext(context.Background()),
				client.API.License.Get.WithPretty(),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				if res.StatusCode == 404 {
					err = forceBasicLicense()
					if err != nil {
						return errors.New("Error when enabled trial license for other tests. You need to check by your hand")
					}

					return nil
				}
			}
		default:
			return errors.New("License is only supported by the elastic library >= v6!")
		}

		return fmt.Errorf("License still exists")
	}

	return nil
}

var testElasticsearchLicense = `
resource "elasticsearch_license" "test" {
  use_basic_license = "true"
}
`

func forceBasicLicense() error {
	meta := testAccProvider.Meta()

	switch meta.(type) {
	case *elastic7.Client:
		client := meta.(*elastic7.Client)
		res, err := client.API.License.PostStartBasic(
			client.API.License.PostStartBasic.WithContext(context.Background()),
			client.API.License.PostStartBasic.WithPretty(),
			client.API.License.PostStartBasic.WithAcknowledge(true),
		)

		if err != nil {
			return err
		}

		if res.IsError() {
			return errors.New("Error when enabled basic license")
		}
	}

	return nil
}
