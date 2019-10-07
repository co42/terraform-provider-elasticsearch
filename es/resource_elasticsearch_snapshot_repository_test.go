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
			return fmt.Errorf("No user ID is set")
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
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

		default:
			return errors.New("Snapshot repository is only supported by the elastic library >= v6!")
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

		switch meta.(type) {
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
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
		default:
			return errors.New("Snapshot repository is only supported by the elastic library >= v6!")
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
