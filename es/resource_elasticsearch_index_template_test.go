package es

import (
	"context"
	"fmt"
	"testing"

	elastic6 "github.com/elastic/go-elasticsearch/v6"
	elastic7 "github.com/elastic/go-elasticsearch/v7"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchIndex(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchIndexTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchIndexTemplate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchIndexTemplateExists("elasticsearch_index_template.test"),
				),
			},
		},
	})
}

func testCheckElasticsearchIndexTemplateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No inde ID is set")
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
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
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
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
		default:
			return errors.New("Index template is only supported by the elastic library >= v6!")
		}

		return nil
	}
}

func testCheckElasticsearchIndexTemplateDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_index_template" {
			continue
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
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
				return nil
			}
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
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
				return nil
			}
		default:
			return errors.New("Index template is only supported by the elastic library >= v6!")
		}

		return fmt.Errorf("Index template %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchIndexTemplate = `
resource "elasticsearch_index_template" "test" {
  name = "terraform-test"
  template = <<EOF
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
