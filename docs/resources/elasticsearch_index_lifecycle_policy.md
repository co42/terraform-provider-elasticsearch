# elasticsearch_index_lifecycle_policy Resource Source

This resource permit to manage the index lifecycle policy in Elasticsearch.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/ilm-put-lifecycle.html

***Supported Elasticsearch version:***
  - v6
  - v7

## Example Usage

It will create ILM policy.

```tf
resource elasticsearch_index_lifecycle_policy "test" {
  name = "terraform-test"
  policy = <<EOF
{
  "policy": {
    "phases": {
      "warm": {
        "min_age": "10d",
        "actions": {
          "forcemerge": {
            "max_num_segments": 1
          }
        }
      },
      "delete": {
        "min_age": "30d",
        "actions": {
          "delete": {}
        }
      }
    }
  }
}
EOF
}
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) Identifier for the policy.
  - **policy**: (required) The policy specification. It's a string as JSON object.

## Attribute Reference

NA