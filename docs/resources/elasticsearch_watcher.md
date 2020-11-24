# elasticsearch_watcher Resource Source

This resource permit to manage watcher in Elasticsearch
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/watcher-api-put-watch.html

***Supported Elasticsearch version:***
  - v6
  - v7

## Example Usage

It will create watcher job.

```tf
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
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) Identifier for the watcher.
  - **trigger**: (optional) The trigger that defines when the watch should run. It's a string as JSON object.
  - **input**: (optional) The input that defines the input that loads the data for the watch. It's a string as JSON object.
  - **condition**: (optional) The condition that defines if the actions should be run. It's a string as JSON object.
  - **actions**: (optional) The list of actions that will be run if the condition matches. It's a string as JSOn object.
  - **throttle_period**: (optional) The minimum time between actions being run.
  - **metadata**: (optional) Metadata json that will be copied into the history entries. It's a string as JSON object.

## Attribute Reference

NA