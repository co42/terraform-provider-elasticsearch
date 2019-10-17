package es

import (
	"context"
	"fmt"
	"testing"

	elastic "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchSecurityUser(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSecurityUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSecurityUser,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityUserExists("elasticsearch_user.test"),
				),
			},
			{
				Config: testElasticsearchSecurityUserUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityUserExists("elasticsearch_user.test"),
				),
			},
			{
				ResourceName:            "elasticsearch_user.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata", "password", "password_hash"},
			},
		},
	})
}

func testCheckElasticsearchSecurityUserExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No user ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Security.GetUser(
			client.API.Security.GetUser.WithContext(context.Background()),
			client.API.Security.GetUser.WithPretty(),
			client.API.Security.GetUser.WithUsername(rs.Primary.ID),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			return errors.Errorf("Error when get user %s: %s", rs.Primary.ID, res.String())
		}

		return nil
	}
}

func testCheckElasticsearchSecurityUserDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_user" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Security.GetUser(
			client.API.Security.GetUser.WithContext(context.Background()),
			client.API.Security.GetUser.WithPretty(),
			client.API.Security.GetUser.WithUsername(rs.Primary.ID),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			if res.StatusCode == 404 {
				return nil
			}
		}

		return fmt.Errorf("User %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchSecurityUser = `
resource "elasticsearch_user" "test" {
  username 	= "terraform-test"
  enabled 	= "true"
  email 	= "no@no.no"
  full_name = "test"
  password 	= "changeme"
  roles 	= ["kibana_user"]
}
`

var testElasticsearchSecurityUserUpdate = `
resource "elasticsearch_user" "test" {
  username 	= "terraform-test"
  enabled 	= "true"
  email 	= "no@no.no"
  full_name = "test2"
  password 	= "changeme2"
  roles 	= ["kibana_user"]
}
`
