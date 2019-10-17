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

func TestAccElasticsearchSnapshotRepository(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSnapshotRepositoryDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSnapshotRepository,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSnapshotRepositoryExists("elasticsearch_snapshot_repository.test"),
				),
			},
			{
				Config: testElasticsearchSnapshotRepositoryUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSnapshotRepositoryExists("elasticsearch_snapshot_repository.test"),
				),
			},
			{
				ResourceName:      "elasticsearch_snapshot_repository.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testCheckElasticsearchSnapshotRepositoryExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No snapshot repository ID is set")
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Snapshot.GetRepository(
			client.API.Snapshot.GetRepository.WithContext(context.Background()),
			client.API.Snapshot.GetRepository.WithPretty(),
			client.API.Snapshot.GetRepository.WithRepository(rs.Primary.ID),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			return errors.Errorf("Error when get snapshot repository %s: %s", rs.Primary.ID, res.String())
		}

		return nil
	}
}

func testCheckElasticsearchSnapshotRepositoryDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_snapshot_repository" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.Snapshot.GetRepository(
			client.API.Snapshot.GetRepository.WithContext(context.Background()),
			client.API.Snapshot.GetRepository.WithPretty(),
			client.API.Snapshot.GetRepository.WithRepository(rs.Primary.ID),
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

		return fmt.Errorf("Snapshot repository %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchSnapshotRepository = `
resource "elasticsearch_snapshot_repository" "test" {
  name		= "terraform-test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
  }
}
`

var testElasticsearchSnapshotRepositoryUpdate = `
resource "elasticsearch_snapshot_repository" "test" {
  name		= "terraform-test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
	"test"	= "test"
  }
}
`
