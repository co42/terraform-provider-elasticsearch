package es

import (
	"context"
	"fmt"
	"testing"

	elastic "github.com/elastic/go-elasticsearch/v8"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchIndexTemplateLegacy(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchIndexTemplateLegacyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchIndexTemplateLegacy,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexTemplateLegacyExists("elasticsearch_index_template_legacy.test"),
				),
			},
			{
				Config: testElasticsearchIndexTemplateLegacyUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexTemplateLegacyExists("elasticsearch_index_template_legacy.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_index_template_legacy.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchIndexTemplateLegacyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No inde ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Indices.GetTemplate(
			client.API.Indices.GetTemplate.WithName(rs.Primary.ID),
			client.API.Indices.GetTemplate.WithContext(context.Background()),
			client.API.Indices.GetTemplate.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			return errors.Errorf("Error when get index template %s: %s", rs.Primary.ID, res.String())
		}

		return nil
	}
}

func testCheckElasticsearchIndexTemplateLegacyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_index_template_legacy" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Indices.DeleteTemplate(
			rs.Primary.ID,
			client.API.Indices.DeleteTemplate.WithContext(context.Background()),
			client.API.Indices.DeleteTemplate.WithPretty(),
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

		return fmt.Errorf("Index template %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchIndexTemplateLegacy = `
resource "elasticsearch_index_template_legacy" "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
  "index_patterns": [
    "test"
  ],
  "settings": {
    "index.refresh_interval": "5s",
	"index.lifecycle.name": "policy-logstash-backup",
    "index.lifecycle.rollover_alias": "logstash-backup-alias"
  },
  "order": 2
}
EOF
}
`

var testElasticsearchIndexTemplateLegacyUpdate = `
resource "elasticsearch_index_template_legacy" "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
  "index_patterns": [
    "test"
  ],
  "settings": {
    "index.refresh_interval": "3s",
	"index.lifecycle.name": "policy-logstash-backup",
    "index.lifecycle.rollover_alias": "logstash-backup-alias"
  },
  "order": 2
}
EOF
}
`
