package es

import (
	"context"
	"fmt"
	"testing"

	elastic6 "github.com/elastic/go-elasticsearch/v6"
	elastic7 "github.com/elastic/go-elasticsearch/v7"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/pkg/errors"
)

func TestAccElasticsearchWatcher(t *testing.T) {

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testCheckElasticsearchWatcherDestroy,
		Steps: []resource.TestStep{
			{
				Config: testElasticsearchWatcher,
				Check: resource.ComposeTestCheckFunc(
					testCheckElasticsearchWatcherExists("elasticsearch_watcher.test"),
				),
			},
			{
				ResourceName:            "elasticsearch_watcher.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"trigger", "input", "condition", "actions", "metadata"},
			},
		},
	})
}

func testCheckElasticsearchWatcherExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No watcher ID is set")
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		// v6
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			res, err := client.API.XPack.WatcherGetWatch(
				rs.Primary.ID,
				client.API.XPack.WatcherGetWatch.WithContext(context.Background()),
				client.API.XPack.WatcherGetWatch.WithPretty(),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				return errors.Errorf("Error when get watcher %s: %s", rs.Primary.ID, res.String())
			}

		// v7
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			res, err := client.API.Watcher.GetWatch(
				rs.Primary.ID,
				client.API.Watcher.GetWatch.WithContext(context.Background()),
				client.API.Watcher.GetWatch.WithPretty(),
			)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			if res.IsError() {
				return errors.Errorf("Error when get watcher %s: %s", rs.Primary.ID, res.String())
			}

		default:
			return errors.New("Watcher is only supported by the elastic library >= v6!")
		}

		return nil
	}
}

func testCheckElasticsearchWatcherDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "elasticsearch_watcher" {
			continue
		}

		meta := testAccProvider.Meta()

		switch meta.(type) {
		// v6
		case *elastic6.Client:
			client := meta.(*elastic6.Client)
			res, err := client.API.XPack.WatcherGetWatch(
				rs.Primary.ID,
				client.API.XPack.WatcherGetWatch.WithContext(context.Background()),
				client.API.XPack.WatcherGetWatch.WithPretty(),
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

		// v7
		case *elastic7.Client:
			client := meta.(*elastic7.Client)
			res, err := client.API.Watcher.GetWatch(
				rs.Primary.ID,
				client.API.Watcher.GetWatch.WithContext(context.Background()),
				client.API.Watcher.GetWatch.WithPretty(),
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
			return errors.New("Watcher is only supported by the elastic library >= v6!")
		}

		return fmt.Errorf("Watcher %q still exists", rs.Primary.ID)
	}

	return nil
}

var testElasticsearchWatcher = `
resource "elasticsearch_watcher" "test" {
  name		= "terraform-test"
  trigger	= <<EOF
{
	"schedule" : { "cron" : "0 0/1 * * * ?" }
}
EOF
  input		= <<EOF
{
	"search" : {
      "request" : {
        "indices" : [
          "logstash*"
        ],
        "body" : {
          "query" : {
            "bool" : {
              "must" : {
                "match": {
                   "response": 404
                }
              },
              "filter" : {
                "range": {
                  "@timestamp": {
                    "from": "{{ctx.trigger.scheduled_time}}||-5m",
                    "to": "{{ctx.trigger.triggered_time}}"
                  }
                }
              }
            }
          }
        }
      }
    }
}
EOF
  condition		= <<EOF
{
	"compare" : { "ctx.payload.hits.total" : { "gt" : 0 }}
}
EOF
  actions		= <<EOF
{
	"email_admin" : {
      "email" : {
        "to" : "admin@domain.host.com",
        "subject" : "404 recently encountered"
      }
    }
}
EOF
}
`
