package es

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	elastic "github.com/elastic/go-elasticsearch/v7"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchSnapshotLifecyclePolicy(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchSnapshotLifecyclePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchSnapshotLifecyclePolicy,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSnapshotLifecyclePolicyExists("elasticsearch_snapshot_lifecycle_policy.test"),
				),
			},
			{
				Config: testElasticsearchSnapshotLifecyclePolicyUpdate,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchSnapshotLifecyclePolicyExists("elasticsearch_snapshot_lifecycle_policy.test"),
				),
			},
			{
				ResourceName:            "elasticsearch_snapshot_lifecycle_policy.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"configs", "retention"},
			},
		},
	})
}

func testCheckElasticsearchSnapshotLifecyclePolicyExists(name string) resource.TestCheckFunc {
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
		res, err := client.API.SlmGetLifecycle(
			client.API.SlmGetLifecycle.WithContext(context.Background()),
			client.API.SlmGetLifecycle.WithPretty(),
			client.API.SlmGetLifecycle.WithPolicyID(rs.Primary.ID),
		)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.IsError() {
			return errors.Errorf("Error when get snapshot lifecycle policy %s: %s", rs.Primary.ID, res.String())
		}

		// Manage Bug https://github.com/elastic/elasticsearch/issues/47664
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		snapshotLifecyclePolicy := make(SnapshotLifecyclePolicy)
		err = json.Unmarshal(b, &snapshotLifecyclePolicy)
		if err != nil {
			return err
		}
		if len(snapshotLifecyclePolicy) == 0 {
			return errors.Errorf("Error when get snapshot lifecycle policy %s: Policy not found", rs.Primary.ID)
		}

		return nil
	}
}

func testCheckElasticsearchSnapshotLifecyclePolicyDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_snapshot_lifecycle_policy" {
			continue
		}

		meta := testAccProvider.Meta()

		client := meta.(*elastic.Client)
		res, err := client.API.SlmGetLifecycle(
			client.API.SlmGetLifecycle.WithContext(context.Background()),
			client.API.SlmGetLifecycle.WithPretty(),
			client.API.SlmGetLifecycle.WithPolicyID(rs.Primary.ID),
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
		// Manage Bug https://github.com/elastic/elasticsearch/issues/47664
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		snapshotLifecyclePolicy := make(SnapshotLifecyclePolicy)
		err = json.Unmarshal(b, &snapshotLifecyclePolicy)
		if err != nil {
			return err
		}
		if len(snapshotLifecyclePolicy) == 0 {
			return nil
		}

		return fmt.Errorf("Snapshot lifecycle policy %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchSnapshotLifecyclePolicy = `

resource "elasticsearch_snapshot_repository" "test" {
  name		= "test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
  }
}

resource "elasticsearch_snapshot_lifecycle_policy" "test" {
  name			= "terraform-test"
  snapshot_name = "<daily-snap-{now/d}>"
  schedule 		= "0 30 1 * * ?"
  repository    = "${elasticsearch_snapshot_repository.test.name}"
  configs		= <<EOF
{
	"indices": ["test-*"],
	"ignore_unavailable": false,
	"include_global_state": false
}
EOF
  retention     = <<EOF
{
    "expire_after": "7d",
    "min_count": 5,
    "max_count": 10
} 
EOF
}
`

var testElasticsearchSnapshotLifecyclePolicyUpdate = `

resource "elasticsearch_snapshot_repository" "test" {
  name		= "test"
  type 		= "fs"
  settings 	= {
	"location" =  "/tmp"
  }
}

resource "elasticsearch_snapshot_lifecycle_policy" "test" {
  name			= "terraform-test"
  snapshot_name = "<daily-snap-{now/d}>"
  schedule 		= "1 30 1 * * ?"
  repository    = "${elasticsearch_snapshot_repository.test.name}"
  configs		= <<EOF
{
	"indices": ["test-*"],
	"ignore_unavailable": false,
	"include_global_state": false
}
EOF
  retention     = <<EOF
{
    "expire_after": "7d",
    "min_count": 5,
    "max_count": 10
} 
EOF
}
`
