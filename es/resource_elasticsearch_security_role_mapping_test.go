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

func TestAccElasticsearchSecurityRoleMapping(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSecurityRoleMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSecurityRoleMapping,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityRoleMappingExists("elasticsearch_role_mapping.test"),
				),
			},
			{
				Config: testElasticsearchSecurityRoleMappingUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSecurityRoleMappingExists("elasticsearch_role_mapping.test"),
				),
			},
			{
				ResourceName:            "elasticsearch_role_mapping.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"metadata", "rules"},
			},
		},
	})
}

func testCheckElasticsearchSecurityRoleMappingExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No role mapping ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Security.GetRoleMapping(
			client.API.Security.GetRoleMapping.WithContext(context.Background()),
			client.API.Security.GetRoleMapping.WithPretty(),
			client.API.Security.GetRoleMapping.WithName(rs.Primary.ID),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			return errors.Errorf("Error when get role mapping %s: %s", rs.Primary.ID, res.String())
		}

		return nil
	}
}

func testCheckElasticsearchSecurityRoleMappingDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_role_mapping" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Security.GetRoleMapping(
			client.API.Security.GetRoleMapping.WithContext(context.Background()),
			client.API.Security.GetRoleMapping.WithPretty(),
			client.API.Security.GetRoleMapping.WithName(rs.Primary.ID),
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

		return fmt.Errorf("Role mapping %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchSecurityRoleMapping = `
resource "elasticsearch_role_mapping" "test" {
  name 		= "terraform-test"
  enabled 	= "true"
  roles 	= ["superuser"]
  rules 	= <<EOF
{
	"field": {
		"groups": "cn=admins,dc=example,dc=com"
	}
}
EOF
}
`

var testElasticsearchSecurityRoleMappingUpdate = `
resource "elasticsearch_role_mapping" "test" {
  name 		= "terraform-test"
  enabled 	= "true"
  roles 	= ["superuser"]
  rules 	= <<EOF
{
	"field": {
		"groups": "cn=admins2,dc=example,dc=com"
	}
}
EOF
}
`
