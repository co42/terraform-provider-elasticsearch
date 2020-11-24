# elasticsearch_snapshot_lifecycle_policy Resource Source

This resource permit to manage snapshot lifecyle policy.
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/slm-api-put.html

***Supported Elasticsearch version:***
  - v7

## Example Usage

It will create SLM policy.

```tf
resource elasticsearch_snapshot_lifecycle_policy "test" {
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
  retention       = <<EOF
{
    "expire_after": "7d",
    "min_count": 5,
    "max_count": 10
}
EOF
}
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) Identifier for the policy.
  - **snapshot_name**: (required) A name automatically given to each snapshot performed by this policy.
  - **schedule**: (required) A periodic or absolute time schedule.
  - **repository**: (required) The snapshot repository that will contain snapshots created by this policy.
  - **configs**: (optional) Configuration for each snapshot that will be created by this policy. It's a string as JSON object.
  - **retention**: (optional) Retention rules used to retain and delete snapshots created by the policy. It's a string as JSON object.

## Attribute Reference

NA