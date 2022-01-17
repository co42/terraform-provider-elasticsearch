# elasticsearch_index_template_legacy Resource Source

This resource permit to manage the index template in Elasticsearch (the legacy API).
You can see the API documentation: https://www.elastic.co/guide/en/elasticsearch/reference/current/indices-templates.html

***Supported Elasticsearch version:***
  - v6
  - v7

## Example Usage

It will create index template.

```tf
resource elasticsearch_index_template_legacy "test" {
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
```

## Argument Reference

***The following arguments are supported:***
  - **name**: (required) Identifier for the template.
  - **template**: (required) The template specification. It's a string as JSON object.

## Attribute Reference

NA