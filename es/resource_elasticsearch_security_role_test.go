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

func TestAccElasticsearchSecurityRole(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSecurityRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSecurityRole,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityRoleExists("elasticsearch_role.test"),
				),
			},
		},
	})
}

func testCheckElasticsearchSecurityRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No role ID is set")
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			res, err := client.API.Security.GetRole(
				client.API.Security.GetRole.WithContext(context.Background()),
				client.API.Security.GetRole.WithPretty(),
				client.API.Security.GetRole.WithName(rs.Primary.ID),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				return errors.Errorf("Error when get security role %s: %s", rs.Primary.ID, res.String())
			}
		default:
			return errors.New("Security role is only supported by the elastic library >= v6!")
		}

		return nil
	}
}

func testCheckElasticsearchSecurityRoleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_role" {
			continue
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			res, err := client.API.Security.GetRole(
				client.API.Security.GetRole.WithContext(context.Background()),
				client.API.Security.GetRole.WithPretty(),
				client.API.Security.GetRole.WithName(rs.Primary.ID),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				if res.StatusCode == 404 {
					return nil
				} else {
					return err
				}

			}
		default:
			return errors.New("Security role is only supported by the elastic library >= v6!")
		}

		return fmt.Errorf("Security role %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchSecurityRole = `
resource "elasticsearch_role" "test" {
  name = "terraform-test"
  indices {
	  names = ["logstash-*"]
	  privileges = ["read"]
  }
  indices {
	  names = ["logstash-*"]
	  privileges = ["read2"]
  }
  cluster = ["all"]
}
`
