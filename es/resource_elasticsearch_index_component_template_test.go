package es

import (
	"context"
	"fmt"
	"testing"

	elastic "github.com/elastic/go-elasticsearch/v7"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchIndexComponentTemplate(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchIndexComponentTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchIndexComponentTemplate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexComponentTemplateExists("elasticsearch_index_component_template.test"),
				),
			},
			{
				Config: testElasticsearchIndexComponentTemplateUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexComponentTemplateExists("elasticsearch_index_component_template.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_index_component_template.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchIndexComponentTemplateExists(name string) resource.TestCheckFunc {
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
		res, err := client.API.Cluster.GetComponentTemplate(
			client.API.Cluster.GetComponentTemplate.WithName(rs.Primary.ID),
			client.API.Cluster.GetComponentTemplate.WithContext(context.Background()),
			client.API.Cluster.GetComponentTemplate.WithPretty(),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			return errors.Errorf("Error when get index component template %s: %s", rs.Primary.ID, res.String())
		}

		return nil
	}
}

func testCheckElasticsearchIndexComponentTemplateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_index_component_template" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Cluster.DeleteComponentTemplate(
			rs.Primary.ID,
			client.API.Cluster.DeleteComponentTemplate.WithContext(context.Background()),
			client.API.Cluster.DeleteComponentTemplate.WithPretty(),
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

		return fmt.Errorf("Index component template %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchIndexComponentTemplate = `
resource "elasticsearch_index_component_template" "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
	"template": {
		"settings": {
			"index.refresh_interval": "5s"
		},
		"mappings": {
			"_source": {
				"enabled": false
			},
			"properties": {
				"host_name": {
					"type": "keyword"
				},
				"created_at": {
					"type": "date",
					"format": "EEE MMM dd HH:mm:ss Z yyyy"
				}
			}
		}
	}
}
EOF
}
`

var testElasticsearchIndexComponentTemplateUpdate = `
resource "elasticsearch_index_component_template" "test" {
  name 		= "terraform-test"
  template 	= <<EOF
{
	"template": {
		"settings": {
			"index.refresh_interval": "3s"
		},
		"mappings": {
			"_source": {
				"enabled": false
			},
			"properties": {
				"host_name": {
					"type": "keyword"
				},
				"created_at": {
					"type": "date",
					"format": "EEE MMM dd HH:mm:ss Z yyyy"
				}
			}
		}
	}
}
EOF
}
`
